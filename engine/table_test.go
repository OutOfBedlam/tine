package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	tb := NewTable()
	tb.AddColumns([]string{"ts", "name", "age"}, []Type{TIME, STRING, INT})
	require.Equal(t, []string{"TS", "NAME", "AGE"}, tb.Columns())
	require.Equal(t, []Type{TIME, STRING, INT}, tb.Types())

	ts := time.Now()
	tb.Add(NewTimeField("ts", ts), NewStringField("name", "foo"), NewIntField("age", 20))
	require.Equal(t, 1, tb.Len())
	require.Equal(t, 1, len(tb.rows))
	require.Equal(t, 3, len(tb.rows[0]))
	require.Equal(t, ts, tb.rows[0][0].Value)
	require.Equal(t, "foo", tb.rows[0][1].Value)
	require.Equal(t, int64(20), tb.rows[0][2].Value)

	tb.Add(NewTimeField("ts", ts), NewStringField("name", "bar"), NewIntField("age", 21))
	require.Equal(t, 2, tb.Len())
	require.Equal(t, []any{ts, ts}, tb.Series("ts"))
	require.Equal(t, []any{"foo", "bar"}, tb.Series("name"))
	require.Equal(t, []any{int64(20), int64(21)}, tb.Series("age"))
	require.Equal(t, []any{ts, "foo", int64(20)}, tb.Row(0))
	require.Equal(t, []any{ts, "bar", int64(21)}, tb.Row(1))
	require.Nil(t, tb.Row(2))

	sel, err := tb.Select([]string{"name"})
	require.NoError(t, err)
	require.Equal(t, []string{"NAME"}, sel.Columns())
	require.Equal(t, []Type{STRING}, sel.Types())
	require.Equal(t, 2, sel.Len())
	require.Equal(t, []any{"foo"}, sel.Row(0))
	require.Equal(t, []any{"bar"}, sel.Row(1))

	sel, err = tb.Select([]string{"name", "age"})
	require.NoError(t, err)
	require.Equal(t, []string{"NAME", "AGE"}, sel.Columns())
	require.Equal(t, []Type{STRING, INT}, sel.Types())
	require.Equal(t, 2, sel.Len())
	require.Equal(t, []any{"foo", int64(20)}, sel.Row(0))
	require.Equal(t, []any{"bar", int64(21)}, sel.Row(1))

	sel, err = tb.Filter(F{"age", GT, 20}).
		Select([]string{"name"})
	require.NoError(t, err)
	require.Equal(t, []string{"NAME"}, sel.Columns())
	require.Equal(t, []Type{STRING}, sel.Types())
	require.Equal(t, 1, sel.Len())
	require.Equal(t, []any{"bar"}, sel.Row(0))

	sel, err = tb.Filter(OR{F{"name", EQ, "foo"}, F{"age", EQ, 21}}).
		Select([]string{"name"})
	require.NoError(t, err)
	require.Equal(t, []string{"NAME"}, sel.Columns())
	require.Equal(t, []Type{STRING}, sel.Types())
	require.Equal(t, 2, sel.Len())
	require.Equal(t, []any{"foo"}, sel.Row(0))
	require.Equal(t, []any{"bar"}, sel.Row(1))

	sel, err = tb.Filter(F{"name", IN, []string{"foo", "bar", "long"}}).
		Select([]string{"name"})
	require.NoError(t, err)
	require.Equal(t, []string{"NAME"}, sel.Columns())
	require.Equal(t, []Type{STRING}, sel.Types())
	require.Equal(t, 2, sel.Len())
	require.Equal(t, []any{"foo"}, sel.Row(0))
	require.Equal(t, []any{"bar"}, sel.Row(1))

	sel, err = tb.Filter(AND{F{"name", EQ, "bar"}, F{"age", GTE, 21}}).
		Select([]string{"age"})
	require.NoError(t, err)
	require.Equal(t, []string{"AGE"}, sel.Columns())
	require.Equal(t, []Type{INT}, sel.Types())
	require.Equal(t, 1, sel.Len())
	require.Equal(t, []any{int64(21)}, sel.Row(0))

	sel = tb.Filter(F{"name", EQ, "foo"}).Compact()
	require.Equal(t, []string{"TS", "NAME", "AGE"}, sel.Columns())
	require.Equal(t, []Type{TIME, STRING, INT}, sel.Types())
	require.Equal(t, 1, sel.Len())
	require.Equal(t, []any{ts, "foo", int64(20)}, sel.Row(0))

	tb.Clear()
	require.Equal(t, 0, len(tb.rows))
}
