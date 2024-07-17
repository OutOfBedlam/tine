package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	tb := NewTable[int64]()
	for i := 0; i < 100; i++ {
		tb.Set(int64(i), NewIntField("v", int64(i)))
	}
	keys := tb.Keys()
	for i := 0; i < 100; i++ {
		require.Equal(t, int64(i), keys[i])
	}

	tb = NewTable[int64]()

	tb.AddColumns([]string{"ts", "name", "age"}, []Type{TIME, STRING, INT})
	require.Equal(t, []string{"TS", "NAME", "AGE"}, tb.Columns())
	require.Equal(t, []Type{TIME, STRING, INT}, tb.Types())

	ts1 := time.Now()
	tb.Set(ts1.Unix(), NewTimeField("ts", ts1), NewStringField("name", "foo"), NewIntField("age", 20))
	require.Equal(t, 1, tb.Len())
	require.Equal(t, 1, len(tb.rows))
	require.Equal(t, 3, len(tb.rows[ts1.Unix()].Fields))
	require.Equal(t, ts1, tb.rows[ts1.Unix()].Fields[0].Value)
	require.Equal(t, "foo", tb.rows[ts1.Unix()].Fields[1].Value)
	require.Equal(t, int64(20), tb.rows[ts1.Unix()].Fields[2].Value)

	ts2 := ts1.Add(1 * time.Second)
	tb.Set(ts2.Unix(), NewTimeField("ts", ts2), NewStringField("name", "bar"), NewIntField("age", 21))
	require.Equal(t, 2, tb.Len())
	require.Equal(t, []any{ts1, ts2}, UnwrapFields(tb.Series("ts")))
	require.Equal(t, []any{"foo", "bar"}, UnwrapFields(tb.Series("name")))
	require.Equal(t, []any{int64(20), int64(21)}, UnwrapFields(tb.Series("age")))
	require.Equal(t, []any{ts1, "foo", int64(20)}, UnwrapFields(tb.Get(ts1.Unix()).Fields))
	require.Equal(t, []any{ts2, "bar", int64(21)}, UnwrapFields(tb.Get(ts2.Unix()).Fields))
	require.Nil(t, tb.Get(12345))

	sel, err := tb.Select([]string{"name"})
	require.NoError(t, err)
	require.Equal(t, []string{"NAME"}, sel.Columns())
	require.Equal(t, []Type{STRING}, sel.Types())
	require.Equal(t, 2, sel.Len())
	require.Equal(t, []any{"foo"}, UnwrapFields(sel.Get(ts1.Unix()).Fields))
	require.Equal(t, []any{"bar"}, UnwrapFields(sel.Get(ts2.Unix()).Fields))

	sel, err = tb.Select([]string{"name", "age"})
	require.NoError(t, err)
	require.Equal(t, []string{"NAME", "AGE"}, sel.Columns())
	require.Equal(t, []Type{STRING, INT}, sel.Types())
	require.Equal(t, 2, sel.Len())
	require.Equal(t, []any{"foo", int64(20)}, UnwrapFields(sel.Get(ts1.Unix()).Fields))
	require.Equal(t, []any{"bar", int64(21)}, UnwrapFields(sel.Get(ts2.Unix()).Fields))

	sel, err = tb.Filter(F{"age", GT, 20}).
		Select([]string{"name"})

	require.NoError(t, err)
	require.Equal(t, []string{"NAME"}, sel.Columns())
	require.Equal(t, []Type{STRING}, sel.Types())
	require.Equal(t, 1, sel.Len())
	require.Equal(t, []any{"bar"}, UnwrapFields(sel.Get(ts2.Unix()).Fields))

	sel, err = tb.Filter(OR{F{"name", EQ, "foo"}, F{"age", EQ, 21}}).
		Select([]string{"name"})

	require.NoError(t, err)
	require.Equal(t, []string{"NAME"}, sel.Columns())
	require.Equal(t, []Type{STRING}, sel.Types())
	require.Equal(t, 2, sel.Len())
	require.Equal(t, []any{"foo"}, UnwrapFields(sel.Get(ts1.Unix()).Fields))
	require.Equal(t, []any{"bar"}, UnwrapFields(sel.Get(ts2.Unix()).Fields))

	sel, err = tb.Filter(F{"name", IN, []string{"foo", "bar", "long"}}).
		Select([]string{"name"})

	require.NoError(t, err)
	require.Equal(t, []string{"NAME"}, sel.Columns())
	require.Equal(t, []Type{STRING}, sel.Types())
	require.Equal(t, 2, sel.Len())
	require.Equal(t, []any{"foo"}, UnwrapFields(sel.Get(ts1.Unix()).Fields))
	require.Equal(t, []any{"bar"}, UnwrapFields(sel.Get(ts2.Unix()).Fields))

	sel, err = tb.Filter(AND{F{"name", EQ, "bar"}, F{"age", GTE, 21}}).
		Select([]string{"age"})

	require.NoError(t, err)
	require.Equal(t, []string{"AGE"}, sel.Columns())
	require.Equal(t, []Type{INT}, sel.Types())
	require.Equal(t, 1, sel.Len())
	require.Equal(t, []any{int64(21)}, UnwrapFields(sel.Get(ts2.Unix()).Fields))

	sel = tb.Filter(F{"name", EQ, "foo"}).Compact()
	require.Equal(t, []string{"TS", "NAME", "AGE"}, sel.Columns())
	require.Equal(t, []Type{TIME, STRING, INT}, sel.Types())
	require.Equal(t, 1, sel.Len())
	require.Equal(t, []any{ts1, "foo", int64(20)}, UnwrapFields(sel.Get(ts1.Unix()).Fields))

	tb.Clear()
	require.Equal(t, 0, len(tb.rows))
}

func TestTableMergeRecords(t *testing.T) {
	tf1 := time.Unix(time.Now().Unix(), 0)
	tf2 := tf1.Add(1 * time.Second)
	tf3 := tf2.Add(1 * time.Second)

	recsA := []Record{
		NewRecord(
			NewStringField("_in", "cpu"),
			NewTimeField("_ts", tf1),
			NewIntField("total", 10),
		),
		NewRecord(
			NewStringField("_in", "cpu"),
			NewTimeField("_ts", tf2),
			NewIntField("total", 20),
		),
	}
	recsB := []Record{
		NewRecord(
			NewStringField("_in", "mem"),
			NewTimeField("_ts", tf1),
			NewIntField("usage", 30),
		),
		NewRecord(
			NewStringField("_in", "mem"),
			NewTimeField("_ts", tf3),
			NewIntField("usage", 40),
		),
	}

	tb := NewTable[int64]()

	for _, rec := range recsA {
		k, _ := rec.Field("_ts").GetTime()
		tb.Set(k.Unix(), rec.Fields()...)
	}
	for _, rec := range recsB {
		k, _ := rec.Field("_ts").GetTime()
		tb.Set(k.Unix(), rec.Fields()...)
	}

	tbC, err := tb.Select([]string{"_in", "_ts", "total", "usage"})
	require.NoError(t, err)

	// for _, rec := range tbC.Rows() {
	// 	t.Log("-->", rec)
	// }

	tbExpect := []Record{
		NewRecord(
			// NewStringField("_in", "cpu"), // TODO how will we deal with over-written _in and _ts?
			NewStringField("_in", "mem"),
			NewTimeField("_ts", tf1),
			NewIntField("total", 10),
			NewIntField("usage", 30),
		),
		NewRecord(
			NewStringField("_in", "cpu"),
			NewTimeField("_ts", tf2),
			NewIntField("total", 20),
			nil,
		),
		NewRecord(
			NewStringField("_in", "mem"),
			NewTimeField("_ts", tf3),
			nil,
			NewIntField("usage", 40),
		),
	}
	result := tbC.Rows()
	require.Equal(t, len(tbExpect), len(result))
	for i, r := range result {
		require.Equal(t, tbExpect[i].Fields(), r)
	}

	tbC = tbC.Filter(F{"_ts", LTE, tf2}).Compact()
	tbExpect = tbExpect[0:2]
	result = tbC.Rows()
	require.Equal(t, len(tbExpect), len(result))
	for i, r := range result {
		require.Equal(t, tbExpect[i].Fields(), r)
	}
}
