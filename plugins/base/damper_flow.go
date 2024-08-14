package base

import (
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

type damperFlow struct {
	buffer         []engine.Record
	bufferSize     int
	bufferLimit    int
	bufferInterval time.Duration
	lastFlush      time.Time
}

func DamperFlow(ctx *engine.Context) engine.Flow {
	bufferSize := ctx.Config().GetInt("buffer_size", 20)
	bufferLimit := ctx.Config().GetInt("buffer_limit", 1000)
	bufferInterval := ctx.Config().GetDuration("interval", 10*time.Second)
	return &damperFlow{
		buffer:         make([]engine.Record, 0, bufferSize),
		bufferLimit:    bufferLimit,
		bufferInterval: bufferInterval,
		lastFlush:      time.Now(),
	}
}

func (df *damperFlow) Open() error      { return nil }
func (df *damperFlow) Close() error     { return nil }
func (df *damperFlow) Parallelism() int { return 1 }

func (df *damperFlow) Process(r []engine.Record, nextFunc engine.FlowNextFunc) {
	df.buffer = append(df.buffer, r...)
	if time.Since(df.lastFlush) >= df.bufferInterval || len(df.buffer) >= df.bufferLimit {
		df.lastFlush = time.Now()
		ret := df.buffer
		if df.bufferSize < len(df.buffer) {
			df.bufferSize = len(df.buffer)
		}
		df.buffer = make([]engine.Record, 0, df.bufferSize)
		nextFunc(ret, nil)
	} else {
		nextFunc(nil, nil)
	}
}
