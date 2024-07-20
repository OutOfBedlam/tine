package engine

import (
	"context"
	"io"
	"log/slog"
	"time"
)

type Context struct {
	base     context.Context
	conf     Config
	pipeline *Pipeline
	logger   *slog.Logger
}

var _ = context.Context((*Context)(nil))

func newContext(pipeline *Pipeline) *Context {
	return &Context{
		base:     context.Background(),
		pipeline: pipeline,
		logger:   slog.Default(),
	}
}

func (ctx *Context) WithConfig(conf Config) *Context {
	return &Context{
		base:     ctx.base,
		pipeline: ctx.pipeline,
		conf:     conf,
		logger:   ctx.logger,
	}
}

func (ctx *Context) WithLogger(logger *slog.Logger) *Context {
	return &Context{
		base:     ctx.base,
		pipeline: ctx.pipeline,
		conf:     ctx.conf,
		logger:   logger,
	}
}

func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.base.Deadline()
}

func (ctx *Context) Config() Config {
	return ctx.conf
}

func (ctx *Context) PipelineName() string {
	return ctx.pipeline.Name
}

func (ctx *Context) Writer() io.Writer {
	return ctx.pipeline.rawWriter
}

func (ctx *Context) SetContentType(contentType string) {
	if ctx.pipeline.setContentTypeFunc != nil && contentType != "" {
		ctx.pipeline.setContentTypeFunc(contentType)
	}
}

func (ctx *Context) SetContentEncoding(contentEncoding string) {
	if ctx.pipeline.setContentEncodingFunc != nil && contentEncoding != "" {
		ctx.pipeline.setContentEncodingFunc(contentEncoding)
	}
}

func (ctx *Context) Done() <-chan struct{} {
	return ctx.base.Done()
}

func (ctx *Context) Err() error {
	return ctx.base.Err()
}

func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.base.Value(key)
}

func (ctx *Context) CircuitBreak() {
	ctx.pipeline.Stop()
}

func (ctx *Context) LogDebug(msg string, keyvals ...interface{}) {
	ctx.logger.Debug(ctx.pipeline.logMsg(msg), keyvals...)
}

func (ctx *Context) LogInfo(msg string, keyvals ...interface{}) {
	ctx.logger.Info(ctx.pipeline.logMsg(msg), keyvals...)
}

func (ctx *Context) LogWarn(msg string, keyvals ...interface{}) {
	ctx.logger.Warn(ctx.pipeline.logMsg(msg), keyvals...)
}
func (ctx *Context) LogError(msg string, keyvals ...interface{}) {
	ctx.logger.Error(ctx.pipeline.logMsg(msg), keyvals...)
}
