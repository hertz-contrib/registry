NODE1=172.18.31.33
REGISTRY=quay.io/coreos/etcd
ETCD_VERSION=latest
# run cluster
prepare-cluster:
	docker-compose up -d

# run single node in docker 
prepare:
	docker pull ${REGISTRY}:${ETCD_VERSION}
	docker run \
  	-p 2379:2379 \
  	-p 2380:2380 \
  	--name etcd ${REGISTRY}:${ETCD_VERSION} \
  	/usr/local/bin/etcd \
  	--data-dir=/etcd-data --name node1 \
  	--initial-advertise-peer-urls http://${NODE1}:2380 --listen-peer-urls http://0.0.0.0:2380 \
  	--advertise-client-urls http://${NODE1}:2379 --listen-client-urls http://0.0.0.0:2379 \
  	--initial-cluster node1=http://${NODE1}:2380