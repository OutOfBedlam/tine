package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"sync/atomic"
	"time"

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
	if interval <= 0 {
		interval = 0
	} else if interval < time.Second {
		interval = time.Second
	}

	ret := &sqliteInlet{SqliteBase: NewBase(ctx)}
	ret.yieldRows = ctx.Config().GetInt("yield_rows", 0)
	ret.countLimit = ctx.Config().GetInt64("count", 0)
	ret.interval = interval
	return ret
}

type sqliteInlet struct {
	*SqliteBase
	yieldRows  int
	interval   time.Duration
	countLimit int64
	runCount   int64
}

var _ = (engine.Inlet)((*sqliteInlet)(nil))

func (si *sqliteInlet) Interval() time.Duration {
	return si.interval
}

func (si *sqliteInlet) Process(next engine.InletNextFunc) {
	runCount := atomic.AddInt64(&si.runCount, 1)
	if si.countLimit > 0 && runCount > si.countLimit {
		next(nil, io.EOF)
		return
	}
	for _, act := range si.Actions {
		doAction(si.DB, si.Ctx, act, next, si.yieldRows)
	}
	if si.countLimit > 0 && runCount >= si.countLimit {
		next(nil, io.EOF)
	}
}

func doAction(db *sql.DB, ctx context.Context, act Action, cb func([]engine.Record, error), yieldRows int) {
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
	rset := []engine.Record{}
	for rows.Next() {
		rec := engine.NewRecord()
		destValues := make([]any, 0, len(colNames))
		for i := 0; i < len(colNames); i++ {
			switch colTypes[i].DatabaseTypeName() {
			case "INTEGER":
				destValues = append(destValues, new(sql.NullInt64))
			case "TEXT":
				destValues = append(destValues, new(sql.NullString))
			case "REAL":
				destValues = append(destValues, new(sql.NullFloat64))
			case "BLOB":
				destValues = append(destValues, new([]byte))
			default:
				cb(nil, fmt.Errorf("unsupported type %s %s for %s",
					colTypes[i].DatabaseTypeName(),
					colTypes[i].ScanType(),
					colNames[i],
				))
				return
			}
		}
		if err := rows.Scan(destValues...); err != nil {
			if c, ok := ctx.(*engine.Context); ok {
				c.LogWarn("sqlite exec", "sql", act.SqlText, "error", err)
			}
			cb(nil, err)
			return
		}
		for i, value := range destValues {
			if nv, ok := value.(*sql.NullInt64); ok {
				if nv.Valid {
					rec = rec.Append(engine.NewField(colNames[i], nv.Int64))
				} else {
					rec = rec.Append(engine.NewFieldWithValue(colNames[i], engine.NewNullValue(engine.INT)))
				}
				continue
			} else if nv, ok := value.(*sql.NullString); ok {
				if nv.Valid {
					rec = rec.Append(engine.NewField(colNames[i], nv.String))
				} else {
					rec = rec.Append(engine.NewFieldWithValue(colNames[i], engine.NewNullValue(engine.STRING)))
				}
				continue
			} else if nv, ok := value.(*sql.NullFloat64); ok {
				if nv.Valid {
					rec = rec.Append(engine.NewField(colNames[i], nv.Float64))
				} else {
					rec = rec.Append(engine.NewFieldWithValue(colNames[i], engine.NewNullValue(engine.FLOAT)))
				}
				continue
			} else if nv, ok := value.(*[]byte); ok {
				rec = rec.Append(engine.NewField(colNames[i], *nv))
				continue
			}
		}
		rset = append(rset, rec)
		if yieldRows > 0 && len(rset) >= yieldRows {
			cb(rset, nil)
			rset = []engine.Record{}
		}
	}
	if len(rset) > 0 {
		cb(rset, nil)
	}
}
