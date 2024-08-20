package sqlite

import (
	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "sqlite",
		Factory: SqliteOutlet,
	})
}

func SqliteOutlet(ctx *engine.Context) engine.Outlet {
	return &sqliteOutlet{NewBase(ctx)}
}

type sqliteOutlet struct {
	*SqliteBase
}

var _ = (engine.Outlet)((*sqliteOutlet)(nil))

func (so *sqliteOutlet) Handle(recs []engine.Record) error {
	for _, rec := range recs {
		for _, act := range so.Actions {
			args := make([]any, len(act.Fields))
			if len(act.Fields) > 0 {
				for i, field := range rec.Fields(act.Fields...) {
					if field == nil {
						continue
					}
					switch field.Type() {
					case engine.TIME:
						// convert to unix epoch
						args[i], _ = field.Value.Int64()
					default:
						args[i] = field.Value.Raw()
					}
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
