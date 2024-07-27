package engine

import (
	"fmt"
	"net/textproto"
)

func CanonicalTagKey(key string) string {
	return textproto.CanonicalMIMEHeaderKey(key)
}

type Tags map[string]*Value

func (t Tags) Get(key string) *Value {
	if t == nil {
		return nil
	}
	return t[key]
}

func (t Tags) Set(key string, value *Value) {
	t[key] = value
}

func (t Tags) Del(key string) {
	delete(t, key)
}

func (t Tags) Keys() []string {
	if t == nil {
		return nil
	}
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	return keys
}

func (t Tags) Clone() Tags {
	if t == nil {
		return nil
	}
	ret := make(Tags)
	for k, v := range t {
		ret[k] = v
	}
	return ret
}

func (t Tags) Clear() {
	for k := range t {
		delete(t, k)
	}
}

func (t Tags) Len() int {
	return len(t)
}

func (t Tags) IsEmpty() bool {
	return len(t) == 0
}

func (t Tags) IsNotEmpty() bool {
	return len(t) > 0
}

func (t Tags) Merge(other Tags) {
	for k, v := range other {
		t[k] = v
	}
}

func (t Tags) MergeWithPrefix(m Tags, prefix string) {
	for k, v := range m {
		t[prefix+k] = v
	}
}

func (t Tags) MergeWithSuffix(m Tags, suffix string) {
	for k, v := range m {
		t[k+suffix] = v
	}
}

func (t Tags) MergeWithPrefixSuffix(m Tags, prefix, suffix string) {
	for k, v := range m {
		t[prefix+k+suffix] = v
	}
}

func GetTagString(t Tags, key string) string {
	if t == nil {
		return ""
	}
	fmt.Println(t.Keys())
	if v := t.Get(key); v != nil {
		if s, ok := v.String(); ok {
			return s
		}
	}
	return ""
}

func GetTagInt64(t Tags, key string) int64 {
	if t == nil {
		return 0
	}
	if v := t.Get(key); v != nil {
		if i, ok := v.Int64(); ok {
			return i
		}
	}
	return 0
}
