package redis_opentracing

import (
	"context"

	"github.com/go-redis/redis/extra/rediscmd"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type RedisTracingHook struct {
	tracer opentracing.Tracer
}

var _ redis.Hook = RedisTracingHook{}

func (hook RedisTracingHook) createSpan(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		childSpan := hook.tracer.StartSpan(operationName, opentracing.ChildOf(span.Context()))
		return childSpan, opentracing.ContextWithSpan(ctx, childSpan)
	}

	return opentracing.StartSpanFromContextWithTracer(ctx, hook.tracer, operationName)
}

func (hook RedisTracingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span, newCtx := hook.createSpan(ctx, cmd.FullName())
	span.SetTag("db.system", "redis")
	span.SetTag("db.statement", rediscmd.CmdString(cmd))

	return newCtx, nil
}

func (hook RedisTracingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span := opentracing.SpanFromContext(ctx)
	defer span.Finish()

	if err := cmd.Err(); err != nil {
		recordError(ctx, span, err)
	}
	return nil
}

func (hook RedisTracingHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	summary, cmdsString := rediscmd.CmdsString(cmds)
	span, newCtx := hook.createSpan(ctx, "pipeline "+summary)

	span.SetTag("db.system", "redis")
	span.SetTag("db.redis.num_cmd", len(cmds))
	span.SetTag("db.statement", cmdsString)

	return newCtx, nil
}

func (hook RedisTracingHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := opentracing.SpanFromContext(ctx)
	defer span.Finish()

	if err := cmds[0].Err(); err != nil {
		recordError(ctx, span, err)
	}
	return nil
}

func recordError(ctx context.Context, span opentracing.Span, err error) {
	if err != redis.Nil {
		span.SetTag(string(ext.Error), true)
		span.SetTag("db.error", err.Error())
	}
}

func NewHook(tracer opentracing.Tracer) redis.Hook {
	return &RedisTracingHook{
		tracer: tracer,
	}
}
