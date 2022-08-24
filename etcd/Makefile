NODE1=172.18.31.33
REGISTRY=bitnami/etcd
ETCD_VERSION=latest
# run cluster
prepare-cluster:
	cd example
	docker-compose up -d

# run single node in docker 
prepare:
	docker pull ${REGISTRY}:${ETCD_VERSION}
	docker network create app-tier --driver bridge
	docker run -d --name Etcd-server \
    --network app-tier \
    --publish 2379:2379 \
    --publish 2380:2380 \
    --env ALLOW_NONE_AUTHENTICATION=yes \
    --env ETCD_ADVERTISE_CLIENT_URLS=http://etcd-server:2379 \
    bitnami/etcd:latest