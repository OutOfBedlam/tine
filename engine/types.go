package engine

import (
	"fmt"
	"strings"
)

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

	Tags() Tags
}

func NewRecord(fields ...*Field) Record {
	return &tagRecord{fields: fields, tags: Tags{}}
}

type tagRecord struct {
	fields []*Field
	tags   Tags
}

var _ = Record((*tagRecord)(nil))

func (r *tagRecord) Tags() Tags {
	return r.tags
}

func (r *tagRecord) Append(fields ...*Field) Record {
	r.fields = append(r.fields, fields...)
	return r
}

func (r *tagRecord) AppendOrReplace(fields ...*Field) Record {
	for _, f := range fields {
		found := false
		for i, old := range r.fields {
			if old == nil {
				continue
			}
			if strings.EqualFold(old.Name, f.Name) {
				r.fields[i] = f
				found = true
				break
			}
		}
		if !found {
			r.fields = append(r.fields, f)
		}
	}
	return r
}

func (r *tagRecord) Field(name string) *Field {
	for _, f := range r.fields {
		if f == nil {
			continue
		}
		if strings.EqualFold(f.Name, name) {
			return f
		}
	}
	return nil
}

func (r *tagRecord) Fields(names ...string) []*Field {
	if len(names) == 0 {
		return r.fields
	}
	ret := make([]*Field, len(names))
	for i, name := range names {
		ret[i] = r.Field(name)
	}
	return ret
}

func (r *tagRecord) FieldAt(index int) *Field {
	if index < 0 || index >= len(r.fields) {
		return nil
	}
	return r.fields[index]
}

func (r *tagRecord) FieldsAt(indexes ...int) []*Field {
	ret := make([]*Field, len(indexes))
	for i, idx := range indexes {
		ret[i] = r.FieldAt(idx)
	}
	return ret
}

func (r *tagRecord) Empty() bool {
	return len(r.fields) == 0
}

func (r *tagRecord) Names() []string {
	ret := make([]string, len(r.fields))
	for i, f := range r.fields {
		if f != nil {
			ret[i] = f.Name
		} else {
			ret[i] = ""
		}
	}
	return ret
}

func NewField[T RawValue](name string, value T) *Field {
	return &Field{Name: name, Value: NewValue(value), Tags: Tags{}}
}

// if name is empty or value is nil, return nil
func NewFieldWithValue(name string, value *Value) *Field {
	if name == "" || value == nil {
		return nil
	}
	return &Field{Name: name, Value: value, Tags: Tags{}}
}

// Field is Value with name and Tags
type Field struct {
	Name  string `json:"name"`
	Value *Value `json:"value"`
	Tags  Tags   `json:"tags,omitempty"`
}

func (f *Field) Type() Type {
	return f.Value.kind
}

func (f *Field) IsNull() bool {
	return f.Value.isNull
}

// Clone returns a deep copy of the field
func (v *Field) Clone() *Field {
	return &Field{Name: v.Name, Value: v.Value.Clone(), Tags: v.Tags.Clone()}
}

// Copy returns a shallow copy of the field with a new name
func (v *Field) Copy(newName string) *Field {
	return &Field{
		Name:  newName,
		Value: v.Value,
		Tags:  v.Tags,
	}
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
		return NewField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) IntField() *Field {
	if v, ok := f.Value.Int64(); ok {
		return NewField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) UintField() *Field {
	if v, ok := f.Value.Uint64(); ok {
		return NewField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) FloatField() *Field {
	if v, ok := f.Value.Float64(); ok {
		return NewField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) StringField() *Field {
	if v, ok := f.Value.String(); ok {
		return NewField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) TimeField() *Field {
	if v, ok := f.Value.Time(); ok {
		return NewField(f.Name, v)
	} else {
		return nil
	}
}

func (f *Field) BinaryField() *Field {
	if v, ok := f.Value.Bytes(); ok {
		return NewField(f.Name, v)
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
