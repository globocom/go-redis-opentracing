package redisopentracing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"

	redisopentracing "github.com/globocom/go-redis-opentracing"
)

func TestHook(t *testing.T) {
	assert := assert.New(t)

	t.Run("create a new hook", func(t *testing.T) {
		// act
		sut := redisopentracing.NewHook(opentracing.NoopTracer{})

		// assert
		assert.NotNil(sut)
	})

	t.Run("starts a span before a command and finishes it after the command", func(t *testing.T) {
		// arrange
		tracer := mocktracer.New()
		sut := redisopentracing.NewHook(tracer)

		cmd := redis.NewStringCmd(context.Background(), "get")

		// act
		ctx, err1 := sut.BeforeProcess(context.Background(), cmd)

		// assert
		assert.Nil(err1)
		assert.Len(tracer.FinishedSpans(), 0)

		// act
		err2 := sut.AfterProcess(ctx, cmd)

		// assert
		assert.Nil(err2)
		assert.Len(tracer.FinishedSpans(), 1)
		span := tracer.FinishedSpans()[0]
		assert.Equal("get", span.OperationName)
		assert.Len(span.Tags(), 1)
		assert.Equal("redis", span.Tags()["db.type"])
		assert.Equal(nil, span.Tags()["db.error"])
		assert.Equal(nil, span.Tags()[string(ext.Error)])
	})

	t.Run("starts a span before a command and finishes it after the command, with error", func(t *testing.T) {
		// arrange
		tracer := mocktracer.New()
		sut := redisopentracing.NewHook(tracer)

		cmd := redis.NewStringCmd(context.Background(), "get")
		cmd.SetErr(errors.New("some error"))

		// act
		ctx, err1 := sut.BeforeProcess(context.Background(), cmd)

		// assert
		assert.Nil(err1)
		assert.Len(tracer.FinishedSpans(), 0)

		// act
		err2 := sut.AfterProcess(ctx, cmd)

		// assert
		assert.Nil(err2)
		assert.Len(tracer.FinishedSpans(), 1)
		span := tracer.FinishedSpans()[0]
		assert.Equal("get", span.OperationName)
		assert.Len(span.Tags(), 3)
		assert.Equal("redis", span.Tags()["db.type"])
		assert.Equal("some error", span.Tags()["db.error"])
		assert.Equal(true, span.Tags()[string(ext.Error)])
	})

	t.Run("starts a span before a pipeline and finishes it after the pipeline", func(t *testing.T) {
		// arrange
		tracer := mocktracer.New()
		sut := redisopentracing.NewHook(tracer)

		cmd1 := redis.NewStringCmd(context.Background(), "get")
		cmd2 := redis.NewStringCmd(context.Background(), "dbsize")
		cmd3 := redis.NewStringCmd(context.Background(), "set x y")
		cmds := []redis.Cmder{cmd1, cmd2, cmd3}

		// act
		ctx, err1 := sut.BeforeProcessPipeline(context.Background(), cmds)

		// assert
		assert.Nil(err1)
		assert.Len(tracer.FinishedSpans(), 0)

		// act
		err2 := sut.AfterProcessPipeline(ctx, cmds)

		// assert
		assert.Nil(err2)
		assert.Len(tracer.FinishedSpans(), 1)

		span := tracer.FinishedSpans()[0]
		assert.Equal("pipeline", span.OperationName)

		assert.Len(span.Tags(), 2)
		assert.Equal("redis", span.Tags()["db.type"])
		assert.Equal(3, span.Tags()["db.redis.num_cmd"])

		assert.Equal(nil, span.Tags()["db.error0"])
		assert.Equal(nil, span.Tags()["db.error1"])
		assert.Equal(nil, span.Tags()["db.error2"])
		assert.Equal(nil, span.Tags()[string(ext.Error)])
	})

	t.Run("starts a span before a pipeline and finishes it after the pipeline, with error", func(t *testing.T) {
		// arrange
		tracer := mocktracer.New()
		sut := redisopentracing.NewHook(tracer)

		cmd1 := redis.NewStringCmd(context.Background(), "get")
		cmd2 := redis.NewStringCmd(context.Background(), "dbsize")
		cmd2.SetErr(errors.New("error 1 in pipeline cmd"))
		cmd3 := redis.NewStringCmd(context.Background(), "set x y")
		cmd3.SetErr(errors.New("error 2 in pipeline cmd"))
		cmds := []redis.Cmder{cmd1, cmd2, cmd3}
		// act
		ctx, err1 := sut.BeforeProcessPipeline(context.Background(), cmds)

		// assert
		assert.Nil(err1)
		assert.Len(tracer.FinishedSpans(), 0)

		// act
		err2 := sut.AfterProcessPipeline(ctx, cmds)

		// assert
		assert.Nil(err2)
		assert.Len(tracer.FinishedSpans(), 1)

		span := tracer.FinishedSpans()[0]
		assert.Equal("pipeline", span.OperationName)

		assert.Len(span.Tags(), 5)
		assert.Equal("redis", span.Tags()["db.type"])
		assert.Equal(3, span.Tags()["db.redis.num_cmd"])

		assert.Equal(nil, span.Tags()["db.error0"])
		assert.Equal("error 1 in pipeline cmd", span.Tags()["db.error1"])
		assert.Equal("error 2 in pipeline cmd", span.Tags()["db.error2"])
		assert.Equal(true, span.Tags()[string(ext.Error)])
	})
}
