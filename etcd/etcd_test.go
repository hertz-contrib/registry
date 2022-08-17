package etcd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func TestEtcd(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	cli.Put(ctx, "ddddd", "dddd")
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	gr, _ := cli.Get(ctx, "ddddd")
	for _, kv := range gr.Kvs {
		fmt.Println(kv.Key, ":", kv.Value)
	}
	cancel()
}
