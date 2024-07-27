package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBoolFeild(t *testing.T) {
	f := NewField("boolean", true)
	require.Equal(t, BOOL, f.Type())
	require.Equal(t, true, f.Value.raw)
	require.Equal(t, "boolean", f.Name)

	str := f.StringField()
	require.Equal(t, STRING, str.Type())
	require.Equal(t, "true", str.Value.raw)
	require.Equal(t, "boolean", str.Name)

	ival := f.IntField()
	require.Equal(t, INT, ival.Type())
	require.Equal(t, int64(1), ival.Value.raw)
	require.Equal(t, "boolean", ival.Name)

	uval := f.UintField()
	require.Equal(t, UINT, uval.Type())
	require.Equal(t, uint64(1), uval.Value.raw)
	require.Equal(t, "boolean", uval.Name)

	fval := f.FloatField()
	require.Equal(t, FLOAT, fval.Type())
	require.Equal(t, float64(1), fval.Value.raw)
	require.Equal(t, "boolean", fval.Name)

	tval := f.TimeField()
	require.Nil(t, tval)

	Bval := f.BinaryField()
	require.Nil(t, Bval)
}

func TestInt(t *testing.T) {
	f := NewField("integer", int64(42))
	require.Equal(t, INT, f.Type())
	require.Equal(t, int64(42), f.Value.raw)
	require.Equal(t, "integer", f.Name)

	str := f.StringField()
	require.Equal(t, STRING, str.Type())
	require.Equal(t, "42", str.Value.raw)
	require.Equal(t, "integer", str.Name)

	bval := f.BoolField()
	require.Equal(t, BOOL, bval.Type())
	require.Equal(t, true, bval.Value.raw)
	require.Equal(t, "integer", bval.Name)

	uval := f.UintField()
	require.Equal(t, UINT, uval.Type())
	require.Equal(t, uint64(42), uval.Value.raw)
	require.Equal(t, "integer", uval.Name)

	fval := f.FloatField()
	require.Equal(t, FLOAT, fval.Type())
	require.Equal(t, float64(42), fval.Value.raw)
	require.Equal(t, "integer", fval.Name)

	tval := f.TimeField()
	require.Equal(t, TIME, tval.Type())
	require.Equal(t, int64(42), tval.Value.raw.(time.Time).Unix())
	require.Equal(t, "integer", tval.Name)

	Bval := f.BinaryField()
	require.Nil(t, Bval)
}

func TestUint(t *testing.T) {
	f := NewField("unsigned", uint64(42))
	require.Equal(t, UINT, f.Type())
	require.Equal(t, uint64(42), f.Value.raw)
	require.Equal(t, "unsigned", f.Name)

	str := f.StringField()
	require.Equal(t, STRING, str.Type())
	require.Equal(t, "42", str.Value.raw)
	require.Equal(t, "unsigned", str.Name)

	bval := f.BoolField()
	require.Equal(t, BOOL, bval.Type())
	require.Equal(t, true, bval.Value.raw)
	require.Equal(t, "unsigned", bval.Name)

	ival := f.IntField()
	require.Equal(t, INT, ival.Type())
	require.Equal(t, int64(42), ival.Value.raw)
	require.Equal(t, "unsigned", ival.Name)

	fval := f.FloatField()
	require.Equal(t, FLOAT, fval.Type())
	require.Equal(t, float64(42), fval.Value.raw)
	require.Equal(t, "unsigned", fval.Name)

	tval := f.TimeField()
	require.Equal(t, TIME, tval.Type())
	require.Equal(t, int64(42), tval.Value.raw.(time.Time).Unix())
	require.Equal(t, "unsigned", tval.Name)

	Bval := f.BinaryField()
	require.Nil(t, Bval)
}

func TestFloat(t *testing.T) {
	f := NewField("float", 42.42)
	require.Equal(t, FLOAT, f.Type())
	require.Equal(t, 42.42, f.Value.raw)
	require.Equal(t, "float", f.Name)

	str := f.StringField()
	require.Equal(t, STRING, str.Type())
	require.Equal(t, "42.42", str.Value.raw)
	require.Equal(t, "float", str.Name)

	bval := f.BoolField()
	require.Equal(t, BOOL, bval.Type())
	require.Equal(t, true, bval.Value.raw)
	require.Equal(t, "float", bval.Name)

	ival := f.IntField()
	require.Equal(t, INT, ival.Type())
	require.Equal(t, int64(42), ival.Value.raw)
	require.Equal(t, "float", ival.Name)

	uval := f.UintField()
	require.Equal(t, UINT, uval.Type())
	require.Equal(t, uint64(42), uval.Value.raw)
	require.Equal(t, "float", uval.Name)

	tval := f.TimeField()
	require.Equal(t, TIME, tval.Type())
	require.Equal(t, int64(42.42*1e9), tval.Value.raw.(time.Time).UnixNano())
	require.Equal(t, "float", tval.Name)

	Bval := f.BinaryField()
	require.Nil(t, Bval)
}

func TestString(t *testing.T) {
	f := NewField("string", "1")
	require.Equal(t, STRING, f.Type())
	require.Equal(t, "1", f.Value.raw)
	require.Equal(t, "string", f.Name)

	bval := f.BoolField()
	require.Equal(t, BOOL, bval.Type())
	require.Equal(t, true, bval.Value.raw)
	require.Equal(t, "string", bval.Name)

	ival := f.IntField()
	require.Equal(t, INT, ival.Type())
	require.Equal(t, int64(1), ival.Value.raw)
	require.Equal(t, "string", ival.Name)

	uval := f.UintField()
	require.Equal(t, UINT, uval.Type())
	require.Equal(t, uint64(1), uval.Value.raw)
	require.Equal(t, "string", uval.Name)

	fval := f.FloatField()
	require.Equal(t, FLOAT, fval.Type())
	require.Equal(t, float64(1), fval.Value.raw)
	require.Equal(t, "string", fval.Name)

	tval := f.TimeField()
	require.Nil(t, tval)

	f = NewField("string", "2024-01-01T00:00:00Z")
	tval = f.TimeField()
	require.Equal(t, TIME, tval.Type())
	require.Equal(t, "2024-01-01T00:00:00Z", tval.Value.raw.(time.Time).Format(time.RFC3339))

	Bval := f.BinaryField()
	require.Equal(t, BINARY, Bval.Type())
	require.Equal(t, "2024-01-01T00:00:00Z", string(Bval.Value.raw.([]byte)))
}

func TestBinary(t *testing.T) {
	bv := NewField("bf", []byte("binary"))
	bv.Tags = Tags{}
	bv.Tags.Set("Content-Type", NewValue("text/plain"))
	bv.Tags.Set("Content-Length", NewValue(int64(len("binary"))))

	require.Equal(t, "text/plain", GetTagString(bv.Tags, "Content-Type"))
	require.Equal(t, int64(len("binary")), GetTagInt64(bv.Tags, "Content-Length"))

	f := NewField("bin", bv.Value.raw.([]byte))
	require.Equal(t, BINARY, f.Type())
	require.Equal(t, "binary", string(f.Value.raw.([]byte)))
	require.Equal(t, "bin", f.Name)

	bval := f.BoolField()
	require.Nil(t, bval)

	ival := f.IntField()
	require.Nil(t, ival)

	uval := f.UintField()
	require.Nil(t, uval)

	fval := f.FloatField()
	require.Nil(t, fval)

	tval := f.TimeField()
	require.Nil(t, tval)

	str := f.StringField()
	require.Equal(t, STRING, str.Type())
	require.Equal(t, "binary", str.Value.raw)
}
