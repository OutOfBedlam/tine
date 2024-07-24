package sqlite

import (
	"github.com/OutOfBedlam/tine/drivers/sqlite3"
	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "sqlite",
		Factory: SqliteOutlet,
	})
}

func SqliteOutlet(ctx *engine.Context) engine.Outlet {
	return &sqliteOutlet{sqlite3.New(ctx)}
}

type sqliteOutlet struct {
	*sqlite3.SqliteBase
}

var _ = (engine.Outlet)((*sqliteOutlet)(nil))

func (so *sqliteOutlet) Handle(recs []engine.Record) error {
	for _, rec := range recs {
		for _, act := range so.Actions {
			args := make([]any, len(act.Fields))
			for i, field := range rec.Fields(act.Fields...) {
				if field == nil {
					continue
				}
				switch field.Type() {
				case engine.TIME:
					// convert to unix epoch
					args[i] = field.IntField().Value
				default:
					args[i] = field.Value
				}
			}
			result, err := so.DB.ExecContext(so.Ctx, act.SqlText, args...)
			if err != nil {
				so.Ctx.LogWarn("sqlite exec", "error", err, "sql", act.SqlText, "args", args)
				return err
			}
			rowsAffected, _ := result.RowsAffected()
			so.Ctx.LogDebug("sqlite exec", "rowsAffected", rowsAffected)
		}
	}
	return nil
}
