package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/go-redis/redis/v8"
)

var _ registry.Registry = (*redisRegistry)(nil)

type redisRegistry struct {
	client *redis.Client
	rctx   *registryContext
	mu     sync.Mutex
	wg     sync.WaitGroup
}

type registryContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRedisRegistry create a redis registry
func NewRedisRegistry(addr string, opts ...Option) registry.Registry {
	redisOpts := &redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	}
	for _, opt := range opts {
		opt(redisOpts)
	}
	rdb := redis.NewClient(redisOpts)
	return &redisRegistry{
		client: rdb,
	}
}

func (r *redisRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	rctx := registryContext{}
	rctx.ctx, rctx.cancel = context.WithCancel(context.Background())
	m := newMentor()
	r.wg.Add(1)
	go m.subscribe(rctx.ctx, info, r)
	r.wg.Wait()
	rdb := r.client
	hash, err := prepareRegistryHash(info)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.rctx = &rctx
	rdb.HSet(rctx.ctx, hash.key, hash.field, hash.value)
	rdb.Expire(rctx.ctx, hash.key, defaultExpireTime)
	rdb.Publish(rctx.ctx, hash.key, fmt.Sprintf("%s-%s-%s", register, info.ServiceName, info.Addr.String()))
	r.mu.Unlock()
	go m.monitorTTL(rctx.ctx, hash, info, r)
	go keepAlive(rctx.ctx, hash, r)
	return nil
}

func (r *redisRegistry) Deregister(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	rctx := r.rctx
	rdb := r.client
	hash, err := prepareRegistryHash(info)
	if err != nil {
		return err
	}
	r.mu.Lock()
	rdb.HDel(rctx.ctx, hash.key, hash.field)
	rdb.Publish(rctx.ctx, hash.key, fmt.Sprintf("%s-%s-%s", deregister, info.ServiceName, info.Addr.String()))
	rctx.cancel()
	r.mu.Unlock()
	return nil
}
