package base

import (
	"fmt"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

type mergeFlow struct {
	ctx             *engine.Context
	table           *engine.Table[int64]
	waitLimit       time.Duration
	joinField       string // joinField should be time.Time type, for now.
	namePrefixField string
}

func MergeFlow(ctx *engine.Context) engine.Flow {
	waitLimit := ctx.Config().GetDuration("wait_limit", 10*time.Second)
	joinField := "_ts"
	namePrefixField := "_in"
	return &mergeFlow{
		ctx:             ctx,
		table:           engine.NewTable[int64](),
		waitLimit:       waitLimit,
		joinField:       joinField,
		namePrefixField: namePrefixField,
	}
}

func (mf *mergeFlow) Open() error      { return nil }
func (mf *mergeFlow) Close() error     { return nil }
func (mf *mergeFlow) Parallelism() int { return 1 }

func (mf *mergeFlow) Flush() []engine.Record {
	ret := []engine.Record{}
	for _, r := range mf.table.Rows() {
		ret = append(ret, engine.NewRecord(r...))
	}
	return ret
}

func (mf *mergeFlow) Process(records []engine.Record) ([]engine.Record, error) {
	for _, rec := range records {
		k := rec.Field(mf.joinField)
		if k == nil {
			continue
		}
		ts, ok := k.Value.Time()
		if !ok {
			continue
		}

		np := rec.Field(mf.namePrefixField)
		if np != nil && !np.IsNull() {
			namePrefix, _ := np.Value.String()
			fields := []*engine.Field{}
			for _, f := range rec.Fields() {
				if strings.EqualFold(f.Name, mf.namePrefixField) {
					continue
				}
				if strings.EqualFold(f.Name, mf.joinField) {
					fields = append(fields, f)
				} else {
					f.Name = fmt.Sprintf("%v.%s", namePrefix, f.Name)
					fields = append(fields, f)
				}
			}
			mf.table.Set(ts.Unix(), fields...)
		} else {
			mf.table.Set(ts.Unix(), rec.Fields()...)
		}
	}

	til := engine.Now().Add(-mf.waitLimit)

	selected, remains := mf.table.Split(engine.F{ColName: mf.joinField, Comparator: engine.LT, Comparando: til})
	mf.table = remains

	ret := []engine.Record{}
	for _, r := range selected.Rows() {
		ret = append(ret, engine.NewRecord(r...))
	}
	return ret, nil
}
