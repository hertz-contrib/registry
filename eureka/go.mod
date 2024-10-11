module github.com/hertz-contrib/registry/eureka

go 1.16

require (
	github.com/bytedance/sonic v1.8.3
	github.com/cloudwego/hertz v0.6.0
	github.com/hudl/fargo v1.4.0
	github.com/stretchr/testify v1.8.2
	github.com/cloudwego-contrib/cwgo-pkg/registry/eureka v0.0.0
)

replace github.com/cloudwego-contrib/cwgo-pkg/registry/eureka => ../../cwgo-pkg-registry/registry/eureka
