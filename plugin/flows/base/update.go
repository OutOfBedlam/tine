package base

import (
	"fmt"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
)

type updateFlow struct {
	ctx         *engine.Context
	fields      []string
	fieldValues []*engine.Value
	fieldNames  []updateNameFunc
	tags        []string
	tagValues   []*engine.Value
	tagNames    []updateNameFunc
}

type updateNameFunc func(string) string

func UpdateFlow(ctx *engine.Context) engine.Flow {
	ret := &updateFlow{ctx: ctx}
	return ret
}

func (sf *updateFlow) Open() error {
	set := sf.ctx.Config().GetConfigSlice("set", nil)
	if len(set) == 0 {
		return fmt.Errorf("update: no set field")
	}
	for _, s := range set {
		aField, hasField := s["field"]
		aTag, hasTag := s["tag"]
		aValue := s["value"]
		aName := s.GetString("name", "")
		aNameFormat := s.GetString("name_format", "")

		if !hasField && !hasTag {
			return fmt.Errorf("update: 'field' or 'tag' must be set")
		}
		if hasField && hasTag {
			return fmt.Errorf("update: 'field' and 'tag' cannot be set at the same time")
		}
		if hasField {
			if str, ok := aField.(string); !ok || str == "" {
				return fmt.Errorf("update: 'field' must be a string")
			} else {
				sf.fields = append(sf.fields, str)
			}
		}
		if hasTag {
			if str, ok := aTag.(string); !ok || str == "" {
				return fmt.Errorf("update: 'tag' must be a string")
			} else {
				sf.tags = append(sf.tags, strings.TrimPrefix(str, "#"))
			}
		}
		var newValue *engine.Value
		var newName updateNameFunc

		if aNameFormat != "" {
			newName = func(org string) string { return fmt.Sprintf(aNameFormat, org) }
		}
		if aName != "" {
			newName = func(org string) string { return aName }
		}

		// TODO: the configuration should support time/binary values
		// if caller wants to replace Time value via pipeline DSL, it is handled as a string for now
		switch v := aValue.(type) {
		case string:
			newValue = engine.NewValue(v)
		case int:
			newValue = engine.NewValue(int64(v))
		case float64:
			newValue = engine.NewValue(v)
		case bool:
			newValue = engine.NewValue(v)
		default:
			if v != nil {
				newValue = engine.NewValue(fmt.Sprintf("%v", v))
			} else {
				newValue = nil
			}
		}

		if hasField {
			sf.fieldNames = append(sf.fieldNames, newName)
			sf.fieldValues = append(sf.fieldValues, newValue)
		} else {
			sf.tagNames = append(sf.tagNames, newName)
			sf.tagValues = append(sf.tagValues, newValue)
		}
	}
	return nil
}

func (sf *updateFlow) Close() error     { return nil }
func (sf *updateFlow) Parallelism() int { return 1 }

func (sf *updateFlow) Process(recs []engine.Record, nextFunc engine.FlowNextFunc) {
	for _, r := range recs {
		for i, f := range r.Fields(sf.fields...) {
			if f == nil {
				continue
			}
			nName := sf.fieldNames[i]
			nValue := sf.fieldValues[i]
			if nName != nil {
				f.Name = nName(f.Name)
			}
			if nValue != nil {
				f.Value = nValue.Clone()
			}
		}
		for i, tag := range sf.tags {
			if tag == "" {
				continue
			}
			nName := sf.tagNames[i]
			nValue := sf.tagValues[i]
			tags := r.Tags()
			if nName != nil {
				oldValue := tags.Get(tag)
				if oldValue != nil {
					tags.Del(tag)
					if nValue != nil {
						tags.Set(nName(tag), nValue.Clone())
					} else {
						tags.Set(nName(tag), oldValue)
					}
				}
			} else if nValue != nil {
				tags.Set(tag, nValue.Clone())
			}
		}
	}
	nextFunc(recs, nil)
}
