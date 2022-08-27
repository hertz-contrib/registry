ZOOKEEPRER_VERSION ?= 3.7.0

prepare:
	docker pull zookeeper:$(ZOOKEEPRER_VERSION)
	docker run --name zookeeper -p 2181:2181 -d zookeeper:$(ZOOKEEPRER_VERSION)
