package base

import (
	"fmt"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

func MergeFlow(ctx *engine.Context) engine.Flow {
	waitLimit := ctx.Config().GetDuration("wait_limit", 10*time.Second)
	return &mergeFlow{
		ctx:           ctx,
		table:         engine.NewTable[int64](),
		waitLimit:     waitLimit,
		joinTag:       engine.TAG_TIMESTAMP,
		namePrefixTag: engine.TAG_INLET,
	}
}

type mergeFlow struct {
	ctx           *engine.Context
	table         *engine.Table[int64]
	waitLimit     time.Duration
	joinTag       string // joinTag should be time.Time type, for now.
	namePrefixTag string
}

var _ = engine.Flow((*mergeFlow)(nil))
var _ = engine.BufferedFlow((*mergeFlow)(nil))

func (mf *mergeFlow) Open() error      { return nil }
func (mf *mergeFlow) Close() error     { return nil }
func (mf *mergeFlow) Parallelism() int { return 1 }

func (mf *mergeFlow) Flush(cb engine.FlowNextFunc) {
	ret := []engine.Record{}
	for _, r := range mf.table.Rows() {
		ret = append(ret, engine.NewRecord(r...))
	}
	cb(ret, nil)
}

func (mf *mergeFlow) Process(records []engine.Record, nextFunc engine.FlowNextFunc) {
	for _, rec := range records {
		var tsValue *engine.Value
		var ts time.Time
		if v := rec.Tags().Get(mf.joinTag); v == nil {
			continue
		} else {
			tsValue = v
			if t, ok := tsValue.Time(); !ok {
				continue
			} else {
				ts = t
			}
		}
		var namePrefix string
		if v := rec.Tags().Get(mf.namePrefixTag); v != nil && !v.IsNull() {
			if s, ok := v.String(); ok {
				namePrefix = s
			}
		}
		if namePrefix != "" {
			fields := []*engine.Field{engine.NewFieldWithValue(mf.joinTag, tsValue)}
			for _, f := range rec.Fields() {
				f.Name = fmt.Sprintf("%v.%s", namePrefix, f.Name)
				fields = append(fields, f)
			}
			mf.table.Set(ts.Unix(), fields...)
		} else {
			mf.table.Set(ts.Unix(), rec.Fields()...)
		}
	}

	til := engine.Now().Add(-mf.waitLimit)

	selected, remains := mf.table.Split(engine.F{ColName: mf.joinTag, Comparator: engine.LT, Comparando: til})
	mf.table = remains

	ret := []engine.Record{}
	for _, r := range selected.Rows() {
		ret = append(ret, engine.NewRecord(r...))
	}
	nextFunc(ret, nil)
}
