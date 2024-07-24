package engine

import (
	"fmt"
	"net/http"
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

	// AppendOrReplace returns a new record with fields appended or replaced if the field name already exists
	AppendOrReplace(...*Field) Record
}

func NewRecord(fields ...*Field) Record {
	return sliceRecord(fields)
}

type sliceRecord []*Field

func (r sliceRecord) Append(fields ...*Field) Record {
	r = append(r, fields...)
	return r
}

func (r sliceRecord) AppendOrReplace(fields ...*Field) Record {
	for _, f := range fields {
		found := false
		for i, old := range r {
			if old == nil {
				continue
			}
			if strings.EqualFold(old.Name, f.Name) {
				r[i] = f
				found = true
				break
			}
		}
		if !found {
			r = append(r, f)
		}
	}
	return r
}

func (r sliceRecord) Field(name string) *Field {
	name = strings.ToUpper(name)
	for _, f := range r {
		if f == nil {
			continue
		}
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
		if f != nil {
			ret[i] = strings.ToUpper(f.Name)
		} else {
			ret[i] = ""
		}
	}
	return ret
}

type Field struct {
	Name  string              `json:"name"`
	Value *Value              `json:"value"`
	Tags  map[string][]string `json:"tags,omitempty"`
}

func (f *Field) Type() Type {
	return f.Value.kind
}

func (f *Field) IsNull() bool {
	return f.Value.isNull
}

// Clone returns a deep copy of the field
func (v *Field) Clone() *Field {
	ret := &Field{Name: v.Name, Value: v.Value.Clone()}
	if v.Tags != nil {
		ret.Tags = make(map[string][]string)
		for k, v := range v.Tags {
			ret.Tags[k] = append([]string(nil), v...)
		}
	}
	return ret
}

// Copy returns a shallow copy of the field with a new name
func (v *Field) Copy(newName string) *Field {
	return &Field{
		Name:  newName,
		Value: v.Value,
		Tags:  v.Tags,
	}
}

func (v *Field) AddTag(key, value string) {
	if v.Tags == nil {
		v.Tags = make(map[string][]string)
	}
	http.Header(v.Tags).Add(key, value)
}

func (v *Field) DelTag(key string) {
	if v.Tags == nil {
		return
	}
	http.Header(v.Tags).Del(key)
}

func (v *Field) GetTag(key string) string {
	if v.Tags == nil {
		return ""
	}
	return http.Header(v.Tags).Get(key)
}

func (v *Field) SetTag(key, value string) {
	if v.Tags == nil {
		v.Tags = make(map[string][]string)
	}
	http.Header(v.Tags).Set(key, value)
}

func (v *Field) GetTagValues(key string) []string {
	if v.Tags == nil {
		return nil
	}
	return http.Header(v.Tags).Values(key)
}

func (v *Field) TagNames() []string {
	if v.Tags == nil {
		return nil
	}
	headers := make([]string, 0, len(v.Tags))
	for key := range v.Tags {
		headers = append(headers, key)
	}
	return headers
}

func UnboxFields(fields []*Field) []any {
	ret := make([]any, len(fields))
	for i, f := range fields {
		ret[i] = f.Value.raw
	}
	return ret
}

func (f *Field) String() string {
	return fmt.Sprintf("%s:%s(%v)", f.Name, string(f.Value.kind), f.Value.raw)
}

func (f *Field) BoolField() *Field {
	if v, ok := f.Value.Bool(); ok {
		return NewBoolField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) IntField() *Field {
	if v, ok := f.Value.Int64(); ok {
		return NewIntField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) UintField() *Field {
	if v, ok := f.Value.Uint64(); ok {
		return NewUintField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) FloatField() *Field {
	if v, ok := f.Value.Float64(); ok {
		return NewFloatField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) StringField() *Field {
	if v, ok := f.Value.String(); ok {
		return NewStringField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) TimeField() *Field {
	if v, ok := f.Value.Time(); ok {
		return NewTimeField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) BinaryField() *Field {
	if v, ok := f.Value.Bytes(); ok {
		return NewBinaryField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) Convert(to Type) *Field {
	if f.Type() == to {
		return f
	}
	switch to {
	case BOOL:
		return f.BoolField()
	case INT:
		return f.IntField()
	case UINT:
		return f.UintField()
	case FLOAT:
		return f.FloatField()
	case STRING:
		return f.StringField()
	case TIME:
		return f.TimeField()
	case BINARY:
		return f.BinaryField()
	}
	return nil
}

func (f *Field) Func(fn func(any) bool) bool {
	return fn(f.Value.raw)
}

type Type byte

const (
	UNTYPED Type = 0
	BOOL    Type = 'b' // bool
	INT     Type = 'i' // int64
	UINT    Type = 'u' // uint64
	FLOAT   Type = 'f' // float64
	STRING  Type = 's' // string
	TIME    Type = 't' // time.Time
	BINARY  Type = 'B' // *BinaryType
)

func (typ Type) String() string {
	switch typ {
	case BOOL:
		return "BOOL"
	case INT:
		return "INT"
	case UINT:
		return "UINT"
	case FLOAT:
		return "FLOAT"
	case STRING:
		return "STRING"
	case TIME:
		return "TIME"
	case BINARY:
		return "BINARY"
	default:
		return "UNTYPED"
	}
}

func NewBoolField(name string, value bool) *Field {
	return &Field{Name: name, Value: NewValue(value)}
}

func NewBoolFieldNull(name string) *Field {
	return &Field{Name: name, Value: NewNullValue(BOOL)}
}

func NewIntField(name string, value int64) *Field {
	return &Field{Name: name, Value: NewValue(value)}
}

func NewIntFieldNull(name string) *Field {
	return &Field{Name: name, Value: NewNullValue(INT)}
}

func NewUintField(name string, value uint64) *Field {
	return &Field{Name: name, Value: NewValue(value)}
}

func NewUintFieldNull(name string) *Field {
	return &Field{Name: name, Value: NewNullValue(UINT)}
}

func NewFloatField(name string, value float64) *Field {
	return &Field{Name: name, Value: NewValue(value)}
}

func NewFloatFieldNull(name string) *Field {
	return &Field{Name: name, Value: NewNullValue(FLOAT)}
}

func NewStringField(name string, value string) *Field {
	return &Field{Name: name, Value: NewValue(value)}
}

func NewStringFieldNull(name string) *Field {
	return &Field{Name: name, Value: NewNullValue(STRING)}
}

func NewTimeField(name string, value time.Time) *Field {
	return &Field{Name: name, Value: NewValue(value)}
}

func NewTimeFieldNull(name string) *Field {
	return &Field{Name: name, Value: NewNullValue(TIME)}
}

func NewBinaryField(name string, value []byte) *Field {
	return &Field{Name: name, Value: NewValue(value)}
}

func NewBinaryFieldNull(name string) *Field {
	return &Field{Name: name, Value: NewNullValue(BINARY)}
}
