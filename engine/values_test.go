package engine

import "testing"

func TestValues(t *testing.T) {
	tests := []struct {
		v      *Value
		kind   Type
		isNull bool
		raw    any
	}{
		{NewValue("string"), STRING, false, "string"},
		{NewValue(int64(123)), INT, false, int64(123)},
		{NewValue(float64(123.45)), FLOAT, false, 123.45},
		{NewNullValue(STRING), STRING, true, nil},
		{NewUntypedNullValue(), UNTYPED, true, nil},
	}

	for _, test := range tests {
		if test.v.Type() != test.kind {
			t.Errorf("Expected %v, but got %v", test.kind, test.v.Type())
		}
		if test.v.IsNull() != test.isNull {
			t.Errorf("Expected %v, but got %v", test.isNull, test.v.IsNull())
		}
		if test.v.Raw() != test.raw {
			t.Errorf("Expected %v, but got %v", test.raw, test.v)
		}
	}
}
