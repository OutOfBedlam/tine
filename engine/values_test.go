package engine

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

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

func ExampleNewValue() {
	var df = DefaultValueFormat()
	var v, nv *Value

	v = NewValue("string")
	nv = NewNullValue(STRING)
	fmt.Println(v.Format(df), nv.Format(df))

	v = NewValue(int64(-123))
	nv = NewNullValue(INT)
	fmt.Println(v.Format(df), nv.Format(df))

	v = NewValue(uint64(123))
	nv = NewNullValue(UINT)
	fmt.Println(v.Format(df), nv.Format(df))

	v = NewValue(float64(3.141592))
	nv = NewNullValue(FLOAT)
	fmt.Println(v.Format(df), nv.Format(df))

	v = NewValue(true)
	nv = NewNullValue(BOOL)
	fmt.Println(v.Format(df), nv.Format(df))

	v = NewValue(time.Unix(1721865054, 0))
	nv = NewNullValue(TIME)
	fmt.Println(v.Format(df), nv.Format(df))

	v = NewValue([]byte("binary"))
	nv = NewNullValue(BINARY)
	fmt.Println(v.Format(df), nv.Format(df))

	// Output:
	// string NULL
	// -123 NULL
	// 123 NULL
	// 3.141592 NULL
	// true NULL
	// 2024-07-25T08:50:54+09:00 NULL
	// BINARY(6 B) NULL
}

func ExampleDefaultValueFormat() {
	tm := time.Unix(1721865054, 0)
	vf := DefaultValueFormat()
	fmt.Println(NewValue(tm).Format(vf), NewValue(3.141592).Format(vf))
	// Output: 2024-07-25T08:50:54+09:00 3.141592
}

func ExampleValueFormat_epoch() {
	tm := time.Unix(1721865054, 0)
	vf := DefaultValueFormat()
	vf.Timeformat = NewTimeformatter("s")
	vf.Decimal = 4
	fmt.Println(NewValue(tm).Format(vf), NewValue(3.141592).Format(vf))
	// Output: 1721865054 3.1416
}

func ExampleValueFormat() {
	tm := time.Unix(1721865054, 0)
	vf := DefaultValueFormat()
	vf.Timeformat = NewTimeformatter("2006-01-02 15:04:05")
	vf.Decimal = 2
	fmt.Println(NewValue(tm).Format(vf), NewValue(3.141592).Format(vf))
	// Output: 2024-07-25 08:50:54 3.14
}

func ExampleValue_BoolValue() {
	sv := NewValue("TRUE")
	bv := sv.BoolValue()
	fmt.Println(bv.Format(DefaultValueFormat()))
	// Output: true
}

func ExampleValue_Bool() {
	sv := NewValue("TRUE")
	bv, ok := sv.Bool()
	fmt.Println(bv, ok)
	// Output: true true
}

func ExampleValue_IntValue() {
	sv := NewValue("123")
	iv := sv.IntValue()
	fmt.Println(iv.Format(DefaultValueFormat()))
	// Output: 123
}

func ExampleValue_Int64() {
	sv := NewValue("123")
	iv, ok := sv.Int64()
	fmt.Println(iv, ok)
	// Output: 123 true
}

func ExampleValue_UintValue() {
	sv := NewValue("123")
	iv := sv.UintValue()
	fmt.Println(iv.Format(DefaultValueFormat()))
	// Output: 123
}

func ExampleValue_Uint64() {
	sv := NewValue("123")
	iv, ok := sv.Uint64()
	fmt.Println(iv, ok)
	// Output: 123 true
}

func ExampleValue_FloatValue() {
	sv := NewValue("3.141592")
	iv := sv.FloatValue()
	fmt.Println(iv.Format(DefaultValueFormat()))
	// Output: 3.141592
}

func ExampleValue_Float64() {
	sv := NewValue("3.141592")
	iv, ok := sv.Float64()
	fmt.Println(iv, ok)
	// Output: 3.141592 true
}

func ExampleValue_StringValue() {
	v := NewValue(123.456)
	sv := v.StringValue()
	fmt.Println(sv.Format(DefaultValueFormat()))
	// Output: 123.456
}

func ExampleValue_String() {
	v := NewValue(123.456)
	sv, ok := v.String()
	fmt.Println(sv, ok)
	// Output: 123.456 true
}

func ExampleValue_TimeValue() {
	v := NewValue(int64(1721865054))
	tv := v.TimeValue()
	fmt.Println(tv.Format(DefaultValueFormat()))
	// Output: 2024-07-25T08:50:54+09:00
}

func ExampleValue_Time() {
	v := NewValue(int64(1721865054))
	tv, ok := v.Time()
	fmt.Println(tv.Unix(), ok)
	// Output: 1721865054 true
}

func ExampleValue_BinaryValue() {
	v := NewValue("binary")
	bv := v.BinaryValue()
	fmt.Println(hex.EncodeToString(bv.Raw().([]byte)))
	// Output: 62696e617279
}

func ExampleValue_Bytes() {
	v := NewValue("binary")
	bv, ok := v.Bytes()
	fmt.Println(hex.EncodeToString(bv), ok)
	// Output: 62696e617279 true
}

func ExampleValue_Eq_string_vs_string() {
	v1 := NewValue("string")
	v2 := NewValue("string")
	fmt.Println(v1.Eq(v2), v1.Eq("string"), v1.Eq("not"))
	// Output: true true false
}

func ExampleValue_Eq_string_vs_float() {
	v1 := NewValue("123.456")
	v2 := NewValue(123.456)
	v3 := v2.StringValue()
	fmt.Println(v1.Eq(v2), v1.Eq(v3), v2.Eq("123.456"))
	// Output: true true true
}

func ExampleValue_Eq_int_vs_float() {
	v1 := NewValue(int64(123))
	v2 := NewValue(123.0)
	fmt.Println(v1.Eq(v2), v1.Eq(123), v1.Eq(123.0), v1.Eq(123.4))
	// Output: true true true false
}

func ExampleValue_Gt() {
	v1 := NewValue(int64(123))
	v2 := NewValue(234.0)
	v3 := NewValue(345.4)
	fmt.Println(v2.Gt(v1), v1.Gt(v3), v3.Gt(v2))
	// Output: true false true
}

func ExampleValue_Gt_bool() {
	v1 := NewValue(true)
	v2 := NewValue(false)
	fmt.Println(v1.Gt(v2))
	// Output: true
}

func ExampleValue_Lt() {
	v1 := NewValue(int64(123))
	v2 := NewValue(234.0)
	v3 := NewValue(345.4)
	fmt.Println(v2.Lt(v1), v1.Lt(v2), v3.Lt(v2), v1.Lt(v3))
	// Output: false true false true
}

func ExampleValue_Lt_bool() {
	v1 := NewValue(true)
	v2 := NewValue(true)
	fmt.Println(v2.Lt(v1))
	// Output: true
}
