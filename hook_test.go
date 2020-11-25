package redis_opentracing_test

import (
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"

	redis_opentracing "github.com/globocom/go-redis-opentracing"
)

func TestHook(t *testing.T) {
	assert := assert.New(t)

	t.Run("create a new hook", func(t *testing.T) {
		// act
		sut := redis_opentracing.NewHook(opentracing.NoopTracer{})

		// assert
		assert.NotNil(sut)
	})
}
