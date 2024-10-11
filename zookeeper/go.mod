module github.com/hertz-contrib/registry/zookeeper

go 1.16

require (
	github.com/cloudwego/hertz v0.6.0
	github.com/go-zookeeper/zk v1.0.3
	github.com/stretchr/testify v1.8.2
	github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper v0.0.0
)

require (
	github.com/bytedance/sonic v1.8.3
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper => ../../cwgo-pkg-registry/registry/zookeeper
