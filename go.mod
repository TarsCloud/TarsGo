module github.com/TarsCloud/TarsGo

go 1.13

require (
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.4.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/google/uuid v1.3.0 => github.com/lbbniu/uuid v1.3.2
