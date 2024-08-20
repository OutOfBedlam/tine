package sqlite

import (
	"database/sql"
	"log/slog"
	"sync"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/mattn/go-sqlite3"
)

type SqliteBase struct {
	Ctx     *engine.Context
	DB      *sql.DB
	Actions []Action
}

func NewBase(ctx *engine.Context) *SqliteBase {
	return &SqliteBase{Ctx: ctx}
}

type Action struct {
	SqlText string
	Fields  []string
}

var sqliteRegisterOnce sync.Once

func sqliteLimitDebug(ctx *engine.Context, conn *sqlite3.SQLiteConn) {
	ctx.LogDebug("sqlite3", "LIMIT_LENGTH", conn.GetLimit(sqlite3.SQLITE_LIMIT_LENGTH))
	ctx.LogDebug("sqlite3", "LIMIT_SQL_LENGTH", conn.GetLimit(sqlite3.SQLITE_LIMIT_SQL_LENGTH))
	ctx.LogDebug("sqlite3", "LIMIT_COLUMN", conn.GetLimit(sqlite3.SQLITE_LIMIT_COLUMN))
	ctx.LogDebug("sqlite3", "LIMIT_EXPR_DEPTH", conn.GetLimit(sqlite3.SQLITE_LIMIT_EXPR_DEPTH))
	ctx.LogDebug("sqlite3", "LIMIT_COMPOUND_SELECT", conn.GetLimit(sqlite3.SQLITE_LIMIT_COMPOUND_SELECT))
	ctx.LogDebug("sqlite3", "LIMIT_VDBE_OP", conn.GetLimit(sqlite3.SQLITE_LIMIT_VDBE_OP))
	ctx.LogDebug("sqlite3", "LIMIT_FUNCTION_ARG", conn.GetLimit(sqlite3.SQLITE_LIMIT_FUNCTION_ARG))
	ctx.LogDebug("sqlite3", "LIMIT_ATTACHED", conn.GetLimit(sqlite3.SQLITE_LIMIT_ATTACHED))
	ctx.LogDebug("sqlite3", "LIMIT_LIKE_PATTERN_LENGTH", conn.GetLimit(sqlite3.SQLITE_LIMIT_LIKE_PATTERN_LENGTH))
	ctx.LogDebug("sqlite3", "LIMIT_VARIABLE_NUMBER", conn.GetLimit(sqlite3.SQLITE_LIMIT_VARIABLE_NUMBER))
	ctx.LogDebug("sqlite3", "LIMIT_TRIGGER_DEPTH", conn.GetLimit(sqlite3.SQLITE_LIMIT_TRIGGER_DEPTH))
	ctx.LogDebug("sqlite3", "LIMIT_WORKER_THREADS", conn.GetLimit(sqlite3.SQLITE_LIMIT_WORKER_THREADS))
}

func (sb *SqliteBase) Open() error {
	path := sb.Ctx.Config().GetString("path", "")
	sqliteRegisterOnce.Do(func() {
		sql.Register("sqlite3_with_tine", &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				if sb.Ctx.LogEnabled(slog.LevelDebug) {
					sqliteLimitDebug(sb.Ctx, conn)
				}
				return nil
			},
		})
	})
	if db, err := sql.Open("sqlite3_with_tine", path); err != nil {
		return err
	} else {
		sb.DB = db
	}

	for _, sqlText := range sb.Ctx.Config().GetStringSlice("inits", []string{}) {
		_, err := sb.DB.ExecContext(sb.Ctx, sqlText)
		if err != nil {
			return err
		}
	}
	// IMPROVE: ctx.Config()["actions"] return [][]any instead of [][]string what we expect
	if list, ok := sb.Ctx.Config()["actions"]; ok {
		for _, actItem := range list.([]interface{}) {
			sqlTextAndFields := make([]string, len(actItem.([]interface{})))
			for i, str := range actItem.([]interface{}) {
				sqlTextAndFields[i] = str.(string)
			}
			if len(sqlTextAndFields) == 0 {
				continue
			}
			sqlText := sqlTextAndFields[0]
			fields := sqlTextAndFields[1:]
			act := Action{SqlText: sqlText, Fields: fields}
			sb.Actions = append(sb.Actions, act)
		}
	}

	return nil
}

func (sb *SqliteBase) Close() error {
	if sb.DB != nil {
		sb.DB.Close()
	}
	return nil
}
