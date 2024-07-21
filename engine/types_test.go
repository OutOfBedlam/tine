package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBoolFeild(t *testing.T) {
	f := NewBoolField("boolean", true)
	require.Equal(t, BOOL, f.Type)
	require.Equal(t, true, f.Value)
	require.Equal(t, "boolean", f.Name)

	str := f.ToString()
	require.Equal(t, STRING, str.Type)
	require.Equal(t, "true", str.Value)
	require.Equal(t, "boolean", str.Name)

	ival := f.ToInt()
	require.Equal(t, INT, ival.Type)
	require.Equal(t, int64(1), ival.Value)
	require.Equal(t, "boolean", ival.Name)

	uval := f.ToUint()
	require.Equal(t, UINT, uval.Type)
	require.Equal(t, uint64(1), uval.Value)
	require.Equal(t, "boolean", uval.Name)

	fval := f.ToFloat()
	require.Equal(t, FLOAT, fval.Type)
	require.Equal(t, float64(1), fval.Value)
	require.Equal(t, "boolean", fval.Name)

	tval := f.ToTime()
	require.Nil(t, tval)

	Bval := f.ToBinary()
	require.Nil(t, Bval)
}

func TestInt(t *testing.T) {
	f := NewIntField("integer", 42)
	require.Equal(t, INT, f.Type)
	require.Equal(t, int64(42), f.Value)
	require.Equal(t, "integer", f.Name)

	str := f.ToString()
	require.Equal(t, STRING, str.Type)
	require.Equal(t, "42", str.Value)
	require.Equal(t, "integer", str.Name)

	bval := f.ToBool()
	require.Equal(t, BOOL, bval.Type)
	require.Equal(t, true, bval.Value)
	require.Equal(t, "integer", bval.Name)

	uval := f.ToUint()
	require.Equal(t, UINT, uval.Type)
	require.Equal(t, uint64(42), uval.Value)
	require.Equal(t, "integer", uval.Name)

	fval := f.ToFloat()
	require.Equal(t, FLOAT, fval.Type)
	require.Equal(t, float64(42), fval.Value)
	require.Equal(t, "integer", fval.Name)

	tval := f.ToTime()
	require.Equal(t, TIME, tval.Type)
	require.Equal(t, int64(42), tval.Value.(time.Time).Unix())
	require.Equal(t, "integer", tval.Name)

	Bval := f.ToBinary()
	require.Nil(t, Bval)
}

func TestUint(t *testing.T) {
	f := NewUintField("unsigned", 42)
	require.Equal(t, UINT, f.Type)
	require.Equal(t, uint64(42), f.Value)
	require.Equal(t, "unsigned", f.Name)

	str := f.ToString()
	require.Equal(t, STRING, str.Type)
	require.Equal(t, "42", str.Value)
	require.Equal(t, "unsigned", str.Name)

	bval := f.ToBool()
	require.Equal(t, BOOL, bval.Type)
	require.Equal(t, true, bval.Value)
	require.Equal(t, "unsigned", bval.Name)

	ival := f.ToInt()
	require.Equal(t, INT, ival.Type)
	require.Equal(t, int64(42), ival.Value)
	require.Equal(t, "unsigned", ival.Name)

	fval := f.ToFloat()
	require.Equal(t, FLOAT, fval.Type)
	require.Equal(t, float64(42), fval.Value)
	require.Equal(t, "unsigned", fval.Name)

	tval := f.ToTime()
	require.Equal(t, TIME, tval.Type)
	require.Equal(t, int64(42), tval.Value.(time.Time).Unix())
	require.Equal(t, "unsigned", tval.Name)

	Bval := f.ToBinary()
	require.Nil(t, Bval)
}

func TestFloat(t *testing.T) {
	f := NewFloatField("float", 42.42)
	require.Equal(t, FLOAT, f.Type)
	require.Equal(t, 42.42, f.Value)
	require.Equal(t, "float", f.Name)

	str := f.ToString()
	require.Equal(t, STRING, str.Type)
	require.Equal(t, "42.42", str.Value)
	require.Equal(t, "float", str.Name)

	bval := f.ToBool()
	require.Equal(t, BOOL, bval.Type)
	require.Equal(t, true, bval.Value)
	require.Equal(t, "float", bval.Name)

	ival := f.ToInt()
	require.Equal(t, INT, ival.Type)
	require.Equal(t, int64(42), ival.Value)
	require.Equal(t, "float", ival.Name)

	uval := f.ToUint()
	require.Equal(t, UINT, uval.Type)
	require.Equal(t, uint64(42), uval.Value)
	require.Equal(t, "float", uval.Name)

	tval := f.ToTime()
	require.Equal(t, TIME, tval.Type)
	require.Equal(t, int64(42.42*1e9), tval.Value.(time.Time).UnixNano())
	require.Equal(t, "float", tval.Name)

	Bval := f.ToBinary()
	require.Nil(t, Bval)
}

func TestString(t *testing.T) {
	f := NewStringField("string", "1")
	require.Equal(t, STRING, f.Type)
	require.Equal(t, "1", f.Value)
	require.Equal(t, "string", f.Name)

	bval := f.ToBool()
	require.Equal(t, BOOL, bval.Type)
	require.Equal(t, true, bval.Value)
	require.Equal(t, "string", bval.Name)

	ival := f.ToInt()
	require.Equal(t, INT, ival.Type)
	require.Equal(t, int64(1), ival.Value)
	require.Equal(t, "string", ival.Name)

	uval := f.ToUint()
	require.Equal(t, UINT, uval.Type)
	require.Equal(t, uint64(1), uval.Value)
	require.Equal(t, "string", uval.Name)

	fval := f.ToFloat()
	require.Equal(t, FLOAT, fval.Type)
	require.Equal(t, float64(1), fval.Value)
	require.Equal(t, "string", fval.Name)

	tval := f.ToTime()
	require.Nil(t, tval)

	f = NewStringField("string", "2024-01-01T00:00:00Z")
	tval = f.ToTime()
	require.Equal(t, TIME, tval.Type)
	require.Equal(t, "2024-01-01T00:00:00Z", tval.Value.(time.Time).Format(time.RFC3339))

	Bval := f.ToBinary()
	require.Equal(t, BINARY, Bval.Type)
	require.Equal(t, "2024-01-01T00:00:00Z", string(Bval.Value.(*BinaryValue).data))
}

func TestBinary(t *testing.T) {
	bv := NewBinaryValue([]byte("binary"))
	bv.SetHeader("Content-Type", "text/plain")
	bv.SetHeader("Content-Length", fmt.Sprintf("%d", len("binary")))

	require.Equal(t, "text/plain", bv.GetHeader("content-type"))
	require.Equal(t, fmt.Sprintf("%d", len("binary")), bv.GetHeader("content-length"))

	f := NewBinaryField("bin", bv)
	require.Equal(t, BINARY, f.Type)
	require.Equal(t, "binary", string(f.Value.(*BinaryValue).data))
	require.Equal(t, "bin", f.Name)

	bval := f.ToBool()
	require.Nil(t, bval)

	ival := f.ToInt()
	require.Nil(t, ival)

	uval := f.ToUint()
	require.Nil(t, uval)

	fval := f.ToFloat()
	require.Nil(t, fval)

	tval := f.ToTime()
	require.Nil(t, tval)

	str := f.ToString()
	require.Equal(t, STRING, str.Type)
	require.Equal(t, "binary", str.Value)
}
