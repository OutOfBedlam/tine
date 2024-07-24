package engine

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/OutOfBedlam/tine/util"
)

type RawValue interface {
	string | bool | []byte | int64 | uint64 | float64 | time.Time
}

var reflectTimeType = reflect.TypeOf(time.Time{})

func NewValue[T RawValue](data T) *Value {
	rType := reflect.TypeOf(data)
	if rType == nil {
		return &Value{isNull: true}
	}
	switch rType.Kind() {
	case reflect.String:
		return &Value{kind: STRING, raw: data}
	case reflect.Bool:
		return &Value{kind: BOOL, raw: data}
	case reflect.Int64:
		return &Value{kind: INT, raw: data}
	case reflect.Float64:
		return &Value{kind: FLOAT, raw: data}
	case reflect.Uint64:
		return &Value{kind: UINT, raw: data}
	case reflect.Struct:
		if rType == reflectTimeType {
			return &Value{kind: TIME, raw: data}
		}
	case reflect.Slice:
		if rType.Elem().Kind() == reflect.Uint8 {
			return &Value{kind: BINARY, raw: data}
		}
	}
	return &Value{raw: data}
}

func NewNullValue(kind Type) *Value {
	return &Value{kind: kind, isNull: true, raw: nil}
}

func NewUntypedNullValue() *Value {
	return &Value{kind: UNTYPED, isNull: true, raw: nil}
}

type Value struct {
	kind   Type
	isNull bool
	raw    any
}

func (v *Value) Type() Type {
	return v.kind
}

func (v *Value) IsNull() bool {
	return v.isNull
}

func (v *Value) Raw() any {
	return v.raw
}

func (v *Value) Clone() *Value {
	ret := &Value{kind: v.kind, isNull: v.isNull, raw: v.raw}
	return ret
}

type FormatOption struct {
	Timeformat *Timeformatter
	Decimal    int
}

func (fo FormatOption) FormatTime(tm time.Time) string {
	return fo.Timeformat.Format(tm)
}

func DefaultFormatOption() FormatOption {
	return FormatOption{
		Timeformat: &Timeformatter{format: time.RFC3339, loc: time.Local},
		Decimal:    -1,
	}
}

func (f *Value) Format(opt FormatOption) string {
	var strVal string
	switch f.Type() {
	case BOOL:
		strVal = strconv.FormatBool(f.Raw().(bool))
	case INT:
		strVal = strconv.FormatInt(f.Raw().(int64), 10)
	case UINT:
		strVal = strconv.FormatUint(f.Raw().(uint64), 10)
	case FLOAT:
		strVal = strconv.FormatFloat(f.Raw().(float64), 'f', opt.Decimal, 64)
	case STRING:
		strVal = f.Raw().(string)
	case TIME:
		strVal = opt.Timeformat.Format(f.Raw().(time.Time))
	case BINARY:
		strVal = fmt.Sprintf("BIN(%s)", util.FormatFileSizeInt(len((f.Raw().([]byte)))))
	default:
		panic("unsupported type-" + string(f.Type()))
	}
	return strVal
}

// BoolValue returns the bool value of the value.
// It will return a NewNullValue(BOOL) if it can not be converted to bool.
func (v *Value) BoolValue() *Value {
	if v.isNull {
		return NewNullValue(BOOL)
	}
	switch v.kind {
	case BOOL:
		return NewValue(v.raw.(bool))
	case INT:
		return NewValue(v.raw.(int64) != 0)
	case UINT:
		return NewValue(v.raw.(uint64) != 0)
	case FLOAT:
		return NewValue(v.raw.(float64) != 0)
	case STRING:
		if v, err := strconv.ParseBool(v.raw.(string)); err == nil {
			return NewValue(v)
		}
	case TIME:
		return NewValue(!v.raw.(time.Time).IsZero())
	}
	return NewNullValue(BOOL)
}

func (val *Value) Bool() (bool, bool) {
	if val.kind == BOOL {
		return val.raw.(bool), true
	}
	if b := val.BoolValue(); b.isNull {
		return false, false
	} else {
		return b.raw.(bool), true
	}
}

// IntValue returns the int value of the value.
// It will return a NewNullValue(INT) if it can not be converted to int.
func (val *Value) IntValue() *Value {
	if val.isNull {
		return NewNullValue(INT)
	}
	switch val.kind {
	case INT:
		return NewValue(val.raw.(int64))
	case BOOL:
		if val.raw.(bool) {
			return NewValue(int64(1))
		} else {
			return NewValue(int64(0))
		}
	case UINT:
		return NewValue(int64(val.raw.(uint64)))
	case FLOAT:
		return NewValue(int64(val.raw.(float64)))
	case STRING:
		if v, err := strconv.ParseInt(val.raw.(string), 10, 64); err == nil {
			return NewValue(v)
		}
	case TIME:
		return NewValue(val.raw.(time.Time).Unix())
	}
	return NewNullValue(INT)
}

func (val *Value) Int64() (int64, bool) {
	if val.kind == INT {
		return val.raw.(int64), true
	}
	if i := val.IntValue(); i.isNull {
		return 0, false
	} else {
		return i.raw.(int64), true
	}
}

// UintValue returns the uint value of the value.
// It will return a NewNullValue(UINT) if it can not be converted to uint.
func (val *Value) UintValue() *Value {
	if val.isNull {
		return NewNullValue(UINT)
	}
	switch val.kind {
	case UINT:
		return NewValue(val.raw.(uint64))
	case BOOL:
		if val.raw.(bool) {
			return NewValue(uint64(1))
		} else {
			return NewValue(uint64(0))
		}
	case INT:
		if val.raw.(int64) >= 0 {
			return NewValue(uint64(val.raw.(int64)))
		}
	case FLOAT:
		if val.raw.(float64) >= 0 {
			return NewValue(uint64(val.raw.(float64)))
		}
	case STRING:
		if v, err := strconv.ParseUint(val.raw.(string), 10, 64); err == nil {
			return NewValue(v)
		}
	case TIME:
		return NewValue(uint64(val.raw.(time.Time).Unix()))
	}
	return NewNullValue(UINT)
}

func (val *Value) Uint64() (uint64, bool) {
	if val.kind == UINT {
		return val.raw.(uint64), true
	}
	if i := val.UintValue(); i.isNull {
		return 0, false
	} else {
		return i.raw.(uint64), true
	}
}

// FloatValue returns the float value of the value.
// It will return a NewNullValue(FLOAT) if it can not be converted to float.
func (val *Value) FloatValue() *Value {
	if val.isNull {
		return NewNullValue(FLOAT)
	}
	switch val.Type() {
	case FLOAT:
		return NewValue(val.raw.(float64))
	case BOOL:
		if val.raw.(bool) {
			return NewValue(float64(1))
		}
		return NewValue(float64(0))
	case INT:
		return NewValue(float64(val.raw.(int64)))
	case UINT:
		return NewValue(float64(val.raw.(uint64)))
	case STRING:
		v, err := strconv.ParseFloat(val.raw.(string), 64)
		if err == nil {
			return NewValue(v)
		}
	case TIME:
		return NewValue(float64(val.raw.(time.Time).Unix()))
	}
	return NewNullValue(FLOAT)
}

func (val *Value) Float64() (float64, bool) {
	if val.kind == FLOAT {
		return val.raw.(float64), true
	}
	if fl := val.FloatValue(); fl.isNull {
		return 0, false
	} else {
		return fl.raw.(float64), true
	}
}

// StringValue returns the string value of the value.
// It will return a NewNullValue(STRING) if it can not be converted to string.
func (f *Value) StringValue() *Value {
	if f.isNull {
		return NewNullValue(STRING)
	}
	switch f.Type() {
	case STRING:
		return NewValue(f.raw.(string))
	case BOOL:
		return NewValue(strconv.FormatBool(f.raw.(bool)))
	case INT:
		return NewValue(strconv.FormatInt(f.raw.(int64), 10))
	case UINT:
		return NewValue(strconv.FormatUint(f.raw.(uint64), 10))
	case FLOAT:
		return NewValue(strconv.FormatFloat(f.raw.(float64), 'f', -1, 64))
	case TIME:
		return NewValue(f.raw.(time.Time).Format(time.RFC3339))
	case BINARY:
		return NewValue(string(f.raw.([]byte)))
	}
	return NewNullValue(STRING)
}

func (val *Value) String() (string, bool) {
	if val.kind == STRING {
		return val.raw.(string), true
	}
	if fl := val.StringValue(); fl.isNull {
		return "", false
	} else {
		return fl.raw.(string), true
	}
}

// TimeValue returns the time value of the value.
// It will return a NewNullValue(TIME) if it can not be converted to time.
func (val *Value) TimeValue() *Value {
	if val.isNull {
		return NewNullValue(TIME)
	}
	switch val.Type() {
	case TIME:
		return NewValue(val.raw.(time.Time))
	case STRING:
		if t, err := time.Parse(time.RFC3339, val.raw.(string)); err == nil {
			return NewValue(t)
		}
	case INT:
		return NewValue(time.Unix(val.raw.(int64), 0))
	case UINT:
		return NewValue(time.Unix(int64(val.raw.(uint64)), 0))
	case FLOAT:
		epoch := int64(val.raw.(float64))
		fract := int64((val.raw.(float64) - float64(epoch)) * 1e9)
		return NewValue(time.Unix(epoch, fract))
	}
	return NewNullValue(TIME)
}

func (val *Value) Time() (time.Time, bool) {
	if val.kind == TIME {
		return val.raw.(time.Time), true
	}
	if fl := val.TimeValue(); fl.isNull {
		return time.Time{}, false
	} else {
		return fl.raw.(time.Time), true
	}
}

// BinaryValue returns the binary value of the value.
// It will return a NewNullValue(BINARY) if it can not be converted to binary.
func (val *Value) BinaryValue() *Value {
	if val.isNull {
		return NewNullValue(BINARY)
	}
	switch val.Type() {
	case STRING:
		return NewValue([]byte(val.raw.(string)))
	case BINARY:
		return NewValue(val.raw.([]byte))
	}
	return NewNullValue(BINARY)
}

func (val *Value) Bytes() ([]byte, bool) {
	if val.kind == BINARY {
		return val.raw.([]byte), true
	}
	if b := val.BinaryValue(); b.isNull {
		return nil, false
	} else {
		return b.raw.([]byte), true
	}
}

// Compare this value with other primitive type
// If other is nil, the result is not defined.
func (val *Value) Eq(other any) bool {
	if other == nil {
		return false
	}
	switch val.Type() {
	case BOOL:
		switch o := other.(type) {
		case bool:
			return val.raw.(bool) == o
		}
	case INT:
		switch o := other.(type) {
		case int:
			return val.raw.(int64) == int64(o)
		case int64:
			return val.raw.(int64) == o
		case float32:
			return val.raw.(int64) == int64(o)
		case float64:
			return val.raw.(int64) == int64(o)
		}
	case UINT:
		switch o := other.(type) {
		case int:
			return val.raw.(uint64) == uint64(o)
		case int64:
			return val.raw.(uint64) == uint64(o)
		case uint:
			return val.raw.(uint64) == uint64(o)
		case uint64:
			return val.raw.(uint64) == o
		case float32:
			return val.raw.(uint64) == uint64(o)
		case float64:
			return val.raw.(uint64) == uint64(o)
		}
	case FLOAT:
		switch o := other.(type) {
		case float64:
			return val.raw.(float64) == o
		case int:
			return val.raw.(float64) == float64(o)
		case int64:
			return val.raw.(float64) == float64(o)
		}
	case STRING:
		switch o := other.(type) {
		case string:
			return val.raw.(string) == o
		}
	case TIME:
		switch o := other.(type) {
		case time.Time:
			return val.raw.(time.Time).Equal(o)
		}
	case BINARY:
		switch o := other.(type) {
		case []byte:
			return bytes.Equal(val.raw.([]byte), o)
		}
	}
	return false
}

// Compare this value with other primitive type
// If other is nil, the result is not defined.
func (val *Value) Gt(other any) bool {
	if other == nil {
		return false
	}
	switch val.Type() {
	case BOOL:
		switch o := other.(type) {
		case bool:
			return val.raw.(bool) && !o
		}
	case INT:
		switch o := other.(type) {
		case int:
			return val.raw.(int64) > int64(o)
		case int64:
			return val.raw.(int64) > o
		case float32:
			return val.raw.(int64) > int64(o)
		case float64:
			return val.raw.(int64) > int64(o)
		}
	case UINT:
		switch o := other.(type) {
		case int:
			return val.raw.(uint64) > uint64(o)
		case int64:
			return val.raw.(uint64) > uint64(o)
		case uint:
			return val.raw.(uint64) > uint64(o)
		case uint64:
			return val.raw.(uint64) > o
		case float32:
			return val.raw.(uint64) > uint64(o)
		case float64:
			return val.raw.(uint64) > uint64(o)
		}
	case FLOAT:
		switch o := other.(type) {
		case float32:
			return val.raw.(float64) > float64(o)
		case float64:
			return val.raw.(float64) > o
		case int:
			return val.raw.(float64) > float64(o)
		case int64:
			return val.raw.(float64) > float64(o)
		}
	case STRING:
		switch o := other.(type) {
		case string:
			return val.raw.(string) > o
		}
	case TIME:
		switch o := other.(type) {
		case time.Time:
			return val.raw.(time.Time).After(o)
		}
	}
	return false
}

// Compare this value with other primitive type
// If other is nil, the result is not defined.
func (val *Value) Lt(other any) bool {
	if other == nil {
		return false
	}
	switch val.Type() {
	case BOOL:
		switch o := other.(type) {
		case bool:
			return !val.raw.(bool) && o
		}
	case INT:
		switch o := other.(type) {
		case int:
			return val.raw.(int64) < int64(o)
		case int64:
			return val.raw.(int64) < o
		case float32:
			return val.raw.(int64) < int64(o)
		case float64:
			return val.raw.(int64) < int64(o)
		}
	case UINT:
		switch o := other.(type) {
		case int:
			return val.raw.(uint64) < uint64(o)
		case int64:
			return val.raw.(uint64) < uint64(o)
		case uint:
			return val.raw.(uint64) < uint64(o)
		case uint64:
			return val.raw.(uint64) < o
		case float32:
			return val.raw.(uint64) < uint64(o)
		case float64:
			return val.raw.(uint64) < uint64(o)
		}
	case FLOAT:
		switch o := other.(type) {
		case float32:
			return val.raw.(float64) < float64(o)
		case float64:
			return val.raw.(float64) < o
		case int:
			return val.raw.(float64) < float64(o)
		case int64:
			return val.raw.(float64) < float64(o)
		}
	case STRING:
		switch o := other.(type) {
		case string:
			return val.raw.(string) < o
		}
	case TIME:
		switch o := other.(type) {
		case time.Time:
			return val.raw.(time.Time).Before(o)
		}
	}
	return false
}

func (val *Value) In(other any) bool {
	if other == nil {
		return false
	}
	switch val.Type() {
	case BOOL:
		if o, ok := other.([]bool); ok {
			for _, v := range o {
				if val.raw.(bool) == v {
					return true
				}
			}
		}
	case INT:
		if o, ok := other.([]int); ok {
			for _, v := range o {
				if val.raw.(int64) == int64(v) {
					return true
				}
			}
		}
	case UINT:
		if o, ok := other.([]uint); ok {
			for _, v := range o {
				if val.raw.(uint64) == uint64(v) {
					return true
				}
			}
		}
	case FLOAT:
		if o, ok := other.([]float64); ok {
			for _, v := range o {
				if val.raw.(float64) == v {
					return true
				}
			}
		}
	case STRING:
		if o, ok := other.([]string); ok {
			for _, v := range o {
				if val.raw.(string) == v {
					return true
				}
			}
		}
	case TIME:
		if o, ok := other.([]time.Time); ok {
			for _, v := range o {
				if val.raw.(time.Time).Equal(v) {
					return true
				}
			}
		}
	}
	return false
}
