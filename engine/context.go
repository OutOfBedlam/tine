package engine

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"time"
)

type Context struct {
	base            context.Context
	conf            Config
	pipeline        *Pipeline
	injectionPoints map[string]InjectFunc
	logger          *slog.Logger
}

type InjectFunc func([]Record) ([]Record, error)

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
		base:            ctx.base,
		conf:            conf,
		pipeline:        ctx.pipeline,
		logger:          ctx.logger,
		injectionPoints: ctx.injectionPoints,
	}
}

func (ctx *Context) WithLogger(logger *slog.Logger) *Context {
	return &Context{
		base:            ctx.base,
		conf:            ctx.conf,
		pipeline:        ctx.pipeline,
		logger:          logger,
		injectionPoints: ctx.injectionPoints,
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

func (ctx *Context) SetContentLength(contentLength int) {
	if ctx.pipeline.setContentLengthFunc != nil && contentLength > 0 {
		ctx.pipeline.setContentLengthFunc(contentLength)
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

func (ctx *Context) GetInject(id string) (InjectFunc, bool) {
	if ctx.injectionPoints == nil {
		return nil, false
	}
	ret, ok := ctx.injectionPoints[id]
	return ret, ok
}

func (ctx *Context) Inject(id string, cb InjectFunc) {
	if ctx.injectionPoints == nil {
		ctx.injectionPoints = make(map[string]InjectFunc)
	}
	ctx.injectionPoints[id] = cb
}

func (ctx *Context) CircuitBreak() {
	// If we do not use goroutine here,
	// It will cause deadlock when outlets are calling ctx.CircuitBreak()
	go func() {
		ctx.pipeline.Stop()
	}()
}

func (ctx *Context) LogDebug(msg string, keyvals ...interface{}) {
	ctx.Log(slog.LevelDebug, msg, keyvals...)
}

func (ctx *Context) LogInfo(msg string, keyvals ...interface{}) {
	ctx.Log(slog.LevelInfo, msg, keyvals...)
}

func (ctx *Context) LogWarn(msg string, keyvals ...interface{}) {
	ctx.Log(slog.LevelWarn, msg, keyvals...)
}

func (ctx *Context) LogError(msg string, keyvals ...interface{}) {
	ctx.Log(slog.LevelError, msg, keyvals...)
}

func (ctx *Context) LogEnabled(level slog.Level) bool {
	return ctx.logger.Handler().Enabled(ctx, level)
}

func (ctx *Context) Log(level slog.Level, msg string, keyvals ...interface{}) {
	handler := ctx.logger.Handler()
	if !handler.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, ctx.pipeline.logMsg(msg), pcs[0])
	r.Add(keyvals...)
	handler.Handle(ctx, r)
}
