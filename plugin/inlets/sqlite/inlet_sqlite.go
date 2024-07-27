package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/OutOfBedlam/tine/drivers/sqlite3"
	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "sqlite",
		Factory: SqliteInlet,
	})
}

func SqliteInlet(ctx *engine.Context) engine.Inlet {
	interval := ctx.Config().GetDuration("interval", 0)
	if interval > 0 {
		countLimit := ctx.Config().GetInt64("count", 0)
		ret := &sqlitePullInlet{SqliteBase: sqlite3.New(ctx)}
		ret.countLimit = countLimit
		ret.interval = interval
		return ret
	} else {
		return &sqlitePushInlet{sqlite3.New(ctx)}
	}
}

// ///////////////
// pull
type sqlitePullInlet struct {
	*sqlite3.SqliteBase
	interval   time.Duration
	countLimit int64
	runCount   int64
}

var _ = (engine.PullInlet)((*sqlitePullInlet)(nil))

func (si *sqlitePullInlet) Interval() time.Duration {
	return time.Second
}

func (si *sqlitePullInlet) Pull() ([]engine.Record, error) {
	runCount := atomic.AddInt64(&si.runCount, 1)
	if si.countLimit > 0 && runCount > si.countLimit {
		return nil, io.EOF
	}
	ret := []engine.Record{}
	var retErr error
	for _, act := range si.Actions {
		doAction(si.DB, si.Ctx, act, func(r []engine.Record, err error) {
			ret = append(ret, r...)
			retErr = err
		})
	}
	if si.countLimit > 0 && runCount >= si.countLimit {
		if retErr == nil {
			retErr = io.EOF
		}
	}
	return ret, retErr
}

// ///////////////
// push
type sqlitePushInlet struct {
	*sqlite3.SqliteBase
}

var _ = (engine.PushInlet)((*sqlitePushInlet)(nil))

func (si *sqlitePushInlet) Push(cb func([]engine.Record, error)) {
	for _, act := range si.Actions {
		doAction(si.DB, si.Ctx, act, cb)
	}
	cb(nil, io.EOF)
}

func doAction(db *sql.DB, ctx context.Context, act sqlite3.Action, cb func([]engine.Record, error)) {
	args := make([]any, len(act.Fields))
	for i, field := range act.Fields {
		args[i] = field
	}
	if c, ok := ctx.(*engine.Context); ok {
		c.LogDebug("sqlite exec", "sql", act.SqlText, "args", args)
	}
	rows, err := db.QueryContext(ctx, act.SqlText, args...)
	if err != nil {
		cb(nil, err)
		return
	}
	defer rows.Close()
	colNames, err := rows.Columns()
	if err != nil {
		cb(nil, err)
		return
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		cb(nil, err)
		return
	}
	if len(colNames) != len(colTypes) {
		cb(nil, fmt.Errorf("column names and types mismatch"))
		return
	}
	for rows.Next() {
		rec := engine.NewRecord()
		destValues := make([]any, 0, len(colNames))
		for i := 0; i < len(colNames); i++ {
			switch colTypes[i].DatabaseTypeName() {
			case "INTEGER":
				destValues = append(destValues, new(int64))
			case "TEXT":
				destValues = append(destValues, new(string))
			case "REAL":
				destValues = append(destValues, new(float64))
			case "BLOB":
				destValues = append(destValues, new([]byte))
			default:
				cb(nil, fmt.Errorf("unsupported type %s for %s",
					colTypes[i].DatabaseTypeName(), colNames[i]))
				return
			}
		}
		if err := rows.Scan(destValues...); err != nil {
			cb(nil, err)
			return
		}
		for i, value := range destValues {
			if value == nil {
				switch colTypes[i].DatabaseTypeName() {
				case "INTEGER":
					rec = rec.Append(engine.NewFieldWithValue(colNames[i], engine.NewNullValue(engine.INT)))
				case "TEXT":
					rec = rec.Append(engine.NewFieldWithValue(colNames[i], engine.NewNullValue(engine.STRING)))
				case "REAL":
					rec = rec.Append(engine.NewFieldWithValue(colNames[i], engine.NewNullValue(engine.FLOAT)))
				case "BLOB":
					rec = rec.Append(engine.NewFieldWithValue(colNames[i], engine.NewNullValue(engine.BINARY)))
				}
				continue
			}
			switch val := value.(type) {
			case *int64:
				rec = rec.Append(engine.NewField(colNames[i], *val))
			case *string:
				rec = rec.Append(engine.NewField(colNames[i], *val))
			case *float64:
				rec = rec.Append(engine.NewField(colNames[i], *val))
			case *[]byte:
				bf := engine.NewField(colNames[i], *val)
				rec = rec.Append(bf)
			}
		}
		cb([]engine.Record{rec}, nil)
	}
}
