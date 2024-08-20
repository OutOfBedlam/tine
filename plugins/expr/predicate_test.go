package expr

import (
	"testing"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/stretchr/testify/require"
)

func TestNewPredicate(t *testing.T) {
	tests := []struct {
		code   string
		txCode string
	}{
		{
			"${ load1 } < 2.5",
			"_load1.Value < 2.5",
		},
	}

	for _, tt := range tests {
		p, err := ExprPredicate(tt.code)
		require.NoError(t, err)
		predicate := p.(*exprPredicate)
		require.Equal(t, tt.txCode, predicate.translatedCode)
	}
}

func TestPredict(t *testing.T) {
	tests := []struct {
		expect bool
		code   string
		record engine.Record
	}{
		{
			true,
			`${load1} < 2.5 // float compare with float`,
			engine.NewRecord(engine.NewField("load1", 2.0)),
		},
		{
			true,
			`${load1} == 2 // int compare with int`,
			engine.NewRecord(engine.NewField("load1", int64(2))),
		},
		{
			false,
			`${load1} == "2" // int compare with string`,
			engine.NewRecord(engine.NewField("load1", 2.1)),
		},
		{
			false,
			`${load1} == 2.1 // string compare with float`,
			engine.NewRecord(engine.NewField("load1", "2.1")),
		},
		{
			true,
			`${a} + ${b} == 3.0 // float compare with float`,
			engine.NewRecord(
				engine.NewField("A", int64(1)),
				engine.NewField("B", int64(2)),
			),
		},
		{
			true,
			`${a} == 1 && ${b} == 2`,
			engine.NewRecord(
				engine.NewField("A", int64(1)),
				engine.NewField("B", int64(2)),
			),
		},
		{
			false,
			`${a} == 1 && ${b} != 2`,
			engine.NewRecord(
				engine.NewField("A", int64(1)),
				engine.NewField("B", int64(2)),
			),
		},
		// OR test
		{
			false,
			`${a} == 1 || ${b} != 2`, // non-exists field 'b'
			engine.NewRecord(
				engine.NewField("A", int64(1)),
			),
		},
		{
			true,
			`${a} == 2 || ${b} == 2`,
			engine.NewRecord(
				engine.NewField("A", int64(1)),
				engine.NewField("B", int64(2)),
			),
		},
	}

	for _, tt := range tests {
		p, err := ExprPredicate(tt.code)
		require.NoError(t, err)
		predicate := p.(*exprPredicate)
		result := predicate.Apply(tt.record)
		if result != tt.expect {
			t.Errorf("expect: %v, got: %v exp:%s", tt.expect, result, tt.code)
		}
		require.Equal(t, tt.expect, result)
	}
}
