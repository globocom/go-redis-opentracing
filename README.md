# go-redis-opentracing

<p>
  <img src="https://img.shields.io/github/workflow/status/globocom/go-redis-opentracing/Go?style=flat-square">
  <a href="https://github.com/globocom/go-redis-opentracing/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/globocom/go-buffer?color=blue&style=flat-square">
  </a>
  <img src="https://img.shields.io/github/go-mod/go-version/globocom/go-redis-opentracing?style=flat-square">
  <a href="https://pkg.go.dev/github.com/globocom/go-redis-opentracing">
    <img src="https://img.shields.io/badge/Go-reference-blue?style=flat-square">
  </a>
</p>


[go-redis](https://github.com/go-redis/redis) hook to collect OpenTracing spans.

There are similar older libs that do not benefit from go-redis newer hooks feature. 
This is heavily inspired by https://github.com/go-redis/redis/blob/master/extra/redisotel/redisotel.go, 
but with support for OpenTracing instead of OpenTelemetry.

Also check out our lib https://github.com/globocom/go-redis-prometheus.

## Installation

    go get github.com/globocom/go-redis-opentracing
    
## Usage

```golang
package main

import (
	redisopentracing "github.com/globocom/go-redis-opentracing"
	"github.com/go-redis/redis/v8"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

func main() {
	cfg := &jaegerConfig.Configuration{
		ServiceName: "my-service-name",
	}
	tracer, _, _ := cfg.NewTracer()

	hook := redisopentracing.NewHook(tracer)

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
	})
	client.AddHook(hook)

	// run redis commands...
}
```

## Note on pipelines

Pipelines generate a single span. For each error that occurs on the pipeline, a tag `db.error<commandIndex>` will be set.

## API stability

The API is unstable at this point, and it might change before `v1.0.0` is released.
