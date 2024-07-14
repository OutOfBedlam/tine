package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type OpenCloser interface {
	Open() error
	Close() error
}

type Record interface {
	// Field retursn field by name
	Field(name string) *Field

	// Fields returns fields in the order of names
	// if names is empty, return all fields
	Fields(names ...string) []*Field

	// FieldAt returns field by index
	FieldAt(index int) *Field

	// FieldsAt return fields in the order of indexes
	FieldsAt(indexes ...int) []*Field

	// Empty returns true if the record has no fields
	Empty() bool

	// Names returns all field names
	Names() []string

	// Append returns a new record with fields appended
	Append(...*Field) Record
}

func NewRecord(fields ...*Field) Record {
	return sliceRecord(fields)
}

type sliceRecord []*Field

func (r sliceRecord) Append(fields ...*Field) Record {
	r = append(r, fields...)
	return r
}

func (r sliceRecord) Field(name string) *Field {
	name = strings.ToUpper(name)
	for _, f := range r {
		if strings.ToUpper(f.Name) == name {
			return f
		}
	}
	return nil
}

func (r sliceRecord) Fields(names ...string) []*Field {
	if len(names) == 0 {
		return r
	}
	ret := make([]*Field, len(names))
	for i, name := range names {
		ret[i] = r.Field(name)
	}
	return ret
}

func (r sliceRecord) FieldAt(index int) *Field {
	if index < 0 || index >= len(r) {
		return nil
	}
	return r[index]
}

func (r sliceRecord) FieldsAt(indexes ...int) []*Field {
	ret := make([]*Field, len(indexes))
	for i, idx := range indexes {
		ret[i] = r.FieldAt(idx)
	}
	return ret
}

func (r sliceRecord) Empty() bool {
	return len(r) == 0
}

func (r sliceRecord) Names() []string {
	ret := make([]string, len(r))
	for i, f := range r {
		ret[i] = strings.ToUpper(f.Name)
	}
	return ret
}

func (f *Field) StringWithFormat(tf *Timeformatter, decimal int) string {
	var strVal string
	switch f.Type {
	case BOOL:
		strVal = strconv.FormatBool(f.Value.(bool))
	case INT:
		strVal = strconv.FormatInt(f.Value.(int64), 10)
	case UINT:
		strVal = strconv.FormatUint(f.Value.(uint64), 10)
	case FLOAT:
		strVal = strconv.FormatFloat(f.Value.(float64), 'f', decimal, 64)
	case STRING:
		strVal = f.Value.(string)
	case TIME:
		strVal = tf.Format(f.Value.(time.Time))
	default:
		panic("unsupported type-" + string(f.Type))
	}
	return strVal
}

type Field struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
	Type  Type   `json:"-"`
}

func (f *Field) String() string {
	return f.Name + ":" + fmt.Sprintf("%s(%v)", string(f.Type), f.Value)
}

func (f *Field) ToBool() *Field {
	if f.Type == BOOL {
		return NewBoolField(f.Name, f.Value.(bool))
	}
	switch f.Type {
	case INT:
		return NewBoolField(f.Name, f.Value.(int64) != 0)
	case UINT:
		return NewBoolField(f.Name, f.Value.(uint64) != 0)
	case FLOAT:
		return NewBoolField(f.Name, f.Value.(float64) != 0)
	case STRING:
		if v, err := strconv.ParseBool(f.Value.(string)); err == nil {
			return NewBoolField(f.Name, v)
		}
	case TIME:
		return NewBoolField(f.Name, !f.Value.(time.Time).IsZero())
	}
	return nil
}

func (f *Field) ToInt() *Field {
	if f.Type == INT {
		return NewIntField(f.Name, f.Value.(int64))
	}
	switch f.Type {
	case BOOL:
		if f.Value.(bool) {
			return NewIntField(f.Name, 1)
		} else {
			return NewIntField(f.Name, 0)
		}
	case UINT:
		return NewIntField(f.Name, int64(f.Value.(uint64)))
	case FLOAT:
		return NewIntField(f.Name, int64(f.Value.(float64)))
	case STRING:
		if v, err := strconv.ParseInt(f.Value.(string), 10, 64); err == nil {
			return NewIntField(f.Name, v)
		}
	case TIME:
		return NewIntField(f.Name, f.Value.(time.Time).Unix())
	}
	return nil
}

func (f *Field) ToUint() *Field {
	if f.Type == UINT {
		return NewUintField(f.Name, f.Value.(uint64))
	}
	switch f.Type {
	case BOOL:
		if f.Value.(bool) {
			return NewUintField(f.Name, 1)
		} else {
			return NewUintField(f.Name, 0)
		}
	case INT:
		if f.Value.(int64) >= 0 {
			return NewUintField(f.Name, uint64(f.Value.(int64)))
		}
	case FLOAT:
		if f.Value.(float64) >= 0 {
			return NewUintField(f.Name, uint64(f.Value.(float64)))
		}
	case STRING:
		if v, err := strconv.ParseUint(f.Value.(string), 10, 64); err == nil {
			return NewUintField(f.Name, v)
		}
	case TIME:
		return NewUintField(f.Name, uint64(f.Value.(time.Time).Unix()))
	}
	return nil
}

func (f *Field) ToFloat() *Field {
	if f.Type == FLOAT {
		return NewFloatField(f.Name, f.Value.(float64))
	}
	switch f.Type {
	case BOOL:
		if f.Value.(bool) {
			return NewFloatField(f.Name, 1)
		}
		return NewFloatField(f.Name, 0)
	case INT:
		return NewFloatField(f.Name, float64(f.Value.(int64)))
	case UINT:
		return NewFloatField(f.Name, float64(f.Value.(uint64)))
	case STRING:
		v, err := strconv.ParseFloat(f.Value.(string), 64)
		if err == nil {
			return NewFloatField(f.Name, v)
		}
	case TIME:
		return NewFloatField(f.Name, float64(f.Value.(time.Time).Unix()))
	}
	return nil
}

func (f *Field) ToString() *Field {
	if f.Type == STRING {
		return NewStringField(f.Name, f.Value.(string))
	}
	switch f.Type {
	case BOOL:
		return NewStringField(f.Name, strconv.FormatBool(f.Value.(bool)))
	case INT:
		return NewStringField(f.Name, strconv.FormatInt(f.Value.(int64), 10))
	case UINT:
		return NewStringField(f.Name, strconv.FormatUint(f.Value.(uint64), 10))
	case FLOAT:
		return NewStringField(f.Name, strconv.FormatFloat(f.Value.(float64), 'f', -1, 64))
	case TIME:
		return NewStringField(f.Name, f.Value.(time.Time).Format(time.RFC3339))
	}
	return nil
}

func (f *Field) ToTime() *Field {
	if f.Type == TIME {
		return NewTimeField(f.Name, f.Value.(time.Time))
	}
	switch f.Type {
	case STRING:
		if t, err := time.Parse(time.RFC3339, f.Value.(string)); err == nil {
			return NewTimeField(f.Name, t)
		}
	case INT:
		return NewTimeField(f.Name, time.Unix(f.Value.(int64), 0))
	case UINT:
		return NewTimeField(f.Name, time.Unix(int64(f.Value.(uint64)), 0))
	case FLOAT:
		epoch := int64(f.Value.(float64))
		fract := int64((f.Value.(float64) - float64(epoch)) * 1e9)
		return NewTimeField(f.Name, time.Unix(epoch, fract))
	case BOOL:
		return nil
	}
	return nil
}

func (f *Field) Convert(to Type) *Field {
	if f.Type == to {
		return f
	}
	switch to {
	case BOOL:
		return f.ToBool()
	case INT:
		return f.ToInt()
	case UINT:
		return f.ToUint()
	case FLOAT:
		return f.ToFloat()
	case STRING:
		return f.ToString()
	case TIME:
		return f.ToTime()
	}
	return nil
}

// Compare this field with other primitive type
func (f *Field) Eq(other any) bool {
	if other == nil {
		return false
	}
	switch f.Type {
	case BOOL:
		switch o := other.(type) {
		case bool:
			return f.Value.(bool) == o
		}
	case INT:
		switch o := other.(type) {
		case int:
			return f.Value.(int64) == int64(o)
		case int64:
			return f.Value.(int64) == o
		case float32:
			return f.Value.(int64) == int64(o)
		case float64:
			return f.Value.(int64) == int64(o)
		}
	case UINT:
		switch o := other.(type) {
		case int:
			return f.Value.(uint64) == uint64(o)
		case int64:
			return f.Value.(uint64) == uint64(o)
		case uint:
			return f.Value.(uint64) == uint64(o)
		case uint64:
			return f.Value.(uint64) == o
		case float32:
			return f.Value.(uint64) == uint64(o)
		case float64:
			return f.Value.(uint64) == uint64(o)
		}
	case FLOAT:
		switch o := other.(type) {
		case float64:
			return f.Value.(float64) == o
		case int:
			return f.Value.(float64) == float64(o)
		case int64:
			return f.Value.(float64) == float64(o)
		}
	case STRING:
		switch o := other.(type) {
		case string:
			return f.Value.(string) == o
		}
	case TIME:
		switch o := other.(type) {
		case time.Time:
			return f.Value.(time.Time).Equal(o)
		}
	}
	return false
}

func (f *Field) Gt(other any) bool {
	if other == nil {
		return false
	}
	switch f.Type {
	case BOOL:
		switch o := other.(type) {
		case bool:
			return f.Value.(bool) && !o
		}
	case INT:
		switch o := other.(type) {
		case int:
			return f.Value.(int64) > int64(o)
		case int64:
			return f.Value.(int64) > o
		case float32:
			return f.Value.(int64) > int64(o)
		case float64:
			return f.Value.(int64) > int64(o)
		}
	case UINT:
		switch o := other.(type) {
		case int:
			return f.Value.(uint64) > uint64(o)
		case int64:
			return f.Value.(uint64) > uint64(o)
		case uint:
			return f.Value.(uint64) > uint64(o)
		case uint64:
			return f.Value.(uint64) > o
		case float32:
			return f.Value.(uint64) > uint64(o)
		case float64:
			return f.Value.(uint64) > uint64(o)
		}
	case FLOAT:
		switch o := other.(type) {
		case float32:
			return f.Value.(float64) > float64(o)
		case float64:
			return f.Value.(float64) > o
		case int:
			return f.Value.(float64) > float64(o)
		case int64:
			return f.Value.(float64) > float64(o)
		}
	case STRING:
		switch o := other.(type) {
		case string:
			return f.Value.(string) > o
		}
	case TIME:
		switch o := other.(type) {
		case time.Time:
			return f.Value.(time.Time).After(o)
		}
	}
	return false
}

func (f *Field) Lt(other any) bool {
	if other == nil {
		return false
	}

	switch f.Type {
	case BOOL:
		switch o := other.(type) {
		case bool:
			return !f.Value.(bool) && o
		}
	case INT:
		switch o := other.(type) {
		case int:
			return f.Value.(int64) < int64(o)
		case int64:
			return f.Value.(int64) < o
		case float32:
			return f.Value.(int64) < int64(o)
		case float64:
			return f.Value.(int64) < int64(o)
		}
	case UINT:
		switch o := other.(type) {
		case int:
			return f.Value.(uint64) < uint64(o)
		case int64:
			return f.Value.(uint64) < uint64(o)
		case uint:
			return f.Value.(uint64) < uint64(o)
		case uint64:
			return f.Value.(uint64) < o
		case float32:
			return f.Value.(uint64) < uint64(o)
		case float64:
			return f.Value.(uint64) < uint64(o)
		}
	case FLOAT:
		switch o := other.(type) {
		case float32:
			return f.Value.(float64) < float64(o)
		case float64:
			return f.Value.(float64) < o
		case int:
			return f.Value.(float64) < float64(o)
		case int64:
			return f.Value.(float64) < float64(o)
		}
	case STRING:
		switch o := other.(type) {
		case string:
			return f.Value.(string) < o
		}
	case TIME:
		switch o := other.(type) {
		case time.Time:
			return f.Value.(time.Time).Before(o)
		}
	}
	return false
}

func (f *Field) In(other any) bool {
	if other == nil {
		return false
	}
	switch f.Type {
	case BOOL:
		if o, ok := other.([]bool); ok {
			for _, v := range o {
				if f.Value.(bool) == v {
					return true
				}
			}
		}
	case INT:
		if o, ok := other.([]int); ok {
			for _, v := range o {
				if f.Value.(int64) == int64(v) {
					return true
				}
			}
		}
	case UINT:
		if o, ok := other.([]uint); ok {
			for _, v := range o {
				if f.Value.(uint64) == uint64(v) {
					return true
				}
			}
		}
	case FLOAT:
		if o, ok := other.([]float64); ok {
			for _, v := range o {
				if f.Value.(float64) == v {
					return true
				}
			}
		}
	case STRING:
		if o, ok := other.([]string); ok {
			for _, v := range o {
				if f.Value.(string) == v {
					return true
				}
			}
		}
	case TIME:
		if o, ok := other.([]time.Time); ok {
			for _, v := range o {
				if f.Value.(time.Time).Equal(v) {
					return true
				}
			}
		}
	}
	return false
}

func (f *Field) Func(fn func(any) bool) bool {
	return fn(f.Value)
}

type Type byte

const (
	BOOL   Type = 'b' // bool
	INT    Type = 'i' // int64
	UINT   Type = 'u' // Uint64
	FLOAT  Type = 'f' // float64
	STRING Type = 's' // string
	TIME   Type = 't' // time.Time
)

func NewBoolField(name string, value bool) *Field {
	return &Field{
		Name:  name,
		Value: value,
		Type:  BOOL,
	}
}

func NewIntField(name string, value int64) *Field {
	return &Field{
		Name:  name,
		Value: value,
		Type:  INT,
	}
}

func NewUintField(name string, value uint64) *Field {
	return &Field{
		Name:  name,
		Value: value,
		Type:  UINT,
	}
}

func NewFloatField(name string, value float64) *Field {
	return &Field{
		Name:  name,
		Value: value,
		Type:  FLOAT,
	}
}

func NewStringField(name string, value string) *Field {
	return &Field{
		Name:  name,
		Value: value,
		Type:  STRING,
	}
}

func NewTimeField(name string, value time.Time) *Field {
	return &Field{
		Name:  name,
		Value: value,
		Type:  TIME,
	}
}

func Map2Records(prefix string, obj map[string]any) []Record {
	ret := []Record{}
	for k, v := range obj {
		r := NewRecord()
		subrecs := []Record{}
		switch v := v.(type) {
		case float64:
			r.Append(NewFloatField(prefix+k, v))
		case int:
			r.Append(NewIntField(prefix+k, int64(v)))
		case int64:
			r.Append(NewIntField(prefix+k, v))
		case string:
			r.Append(NewStringField(prefix+k, v))
		case bool:
			r.Append(NewBoolField(prefix+k, v))
		case time.Time:
			r.Append(NewTimeField(prefix+k, v))
		case map[string]any:
			subrecs = append(subrecs, Map2Records(prefix+k+".", v)...)
		case []any:
			// TODO: support array
			continue
		default:
			continue
		}
		ret = append(ret, r)
		ret = append(ret, subrecs...)
	}
	return ret
}

type Timeformatter struct {
	format string
	loc    *time.Location
}

var DefaultTimeformatter = &Timeformatter{format: time.RFC3339, loc: time.Local}

func (tf *Timeformatter) IsEpoch() bool {
	switch tf.format {
	case "ns", "us", "ms", "s":
		return true
	default:
		return false
	}
}

func (tf *Timeformatter) Epoch(t time.Time) int64 {
	switch tf.format {
	case "ns":
		return t.UnixNano()
	case "us":
		return t.UnixNano() / 1e3
	case "ms":
		return t.UnixNano() / 1e6
	case "s":
		return t.Unix()
	default:
		return t.Unix()
	}
}

func (tf *Timeformatter) Format(t time.Time) string {
	switch tf.format {
	case "ns":
		return strconv.FormatInt(t.UnixNano(), 10)
	case "us":
		return strconv.FormatInt(t.UnixNano()/1e3, 10)
	case "ms":
		return strconv.FormatInt(t.UnixNano()/1e6, 10)
	case "s":
		return strconv.FormatInt(t.Unix(), 10)
	default:
		return t.In(tf.loc).Format(tf.format)
	}
}

func (tf *Timeformatter) Parse(str string) (time.Time, error) {
	switch tf.format {
	case "ns":
		ns, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(0, ns), nil
	case "us":
		us, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(0, us*1e3), nil
	case "ms":
		ms, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(0, ms*1e6), nil
	case "s":
		s, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(s, 0), nil
	default:
		return time.Parse(tf.format, str)
	}
}
