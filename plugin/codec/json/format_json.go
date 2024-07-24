package report

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterEncoder(&engine.EncoderReg{
		Name:        "json",
		Factory:     NewJSONEncoder,
		ContentType: "application/json",
	})
}

func NewJSONEncoder(c engine.EncoderConfig) engine.Encoder {
	return &JSONEncoder{
		EncoderConfig: c,
	}
}

type JSONEncoder struct {
	engine.EncoderConfig
	enc *json.Encoder
}

func (jw *JSONEncoder) Encode(recs []engine.Record) error {
	if jw.enc == nil {
		jw.enc = json.NewEncoder(jw.Writer)
		jw.enc.SetIndent(jw.Prefix, jw.Indent)
	}
	switch strings.ToLower(jw.Subformat) {
	default:
		return jw.encodeDefault(recs)
	case "name_time_value":
		return jw.encodeNameTimeValue(recs, false)
	case "time_name_value":
		return jw.encodeNameTimeValue(recs, true)

	}
}

func (jw *JSONEncoder) encodeDefault(recs []engine.Record) error {
	rs := []map[string]any{}
	for _, rec := range recs {
		if rec.Empty() {
			continue
		}
		r := map[string]any{}
		for _, f := range rec.Fields() {
			if f == nil {
				continue
			}
			if f.Type() == engine.TIME {
				ts, _ := f.Value.Time()
				if jw.FormatOption.Timeformat.IsEpoch() {
					r[f.Name] = jw.FormatOption.Timeformat.Epoch(ts)
				} else {
					r[f.Name] = jw.FormatOption.Timeformat.Format(ts)
				}
			} else {
				if f.Type() == engine.BINARY {
					r[f.Name] = "[BINARY]"
				} else {
					r[f.Name] = f.Value.Raw()
				}
			}
		}
		rs = append(rs, r)
	}
	jw.enc.Encode(rs)
	return nil
}

func (jw *JSONEncoder) encodeNameTimeValue(recs []engine.Record, timeFirst bool) error {
	rs := [][]any{}
	for _, rec := range recs {
		var ts time.Time
		if tsf := rec.Field(engine.FIELD_TIMESTAMP); tsf == nil {
			ts = time.Now()
		} else {
			if v, ok := tsf.Value.Time(); ok {
				ts = v
			} else {
				ts = time.Now()
			}
		}
		var timestamp any
		if jw.FormatOption.Timeformat.IsEpoch() {
			timestamp = jw.FormatOption.Timeformat.Epoch(ts)
		} else {
			timestamp = jw.FormatOption.Timeformat.Format(ts)
		}
		var namePrefix string
		if inf := rec.Field(engine.FIELD_INLET); inf != nil {
			if v, ok := inf.Value.String(); ok {
				namePrefix = v + "."
			}
		}
		for _, field := range rec.Fields(jw.Fields...) {
			if field == nil || field.Name == engine.FIELD_TIMESTAMP || field.Name == engine.FIELD_INLET {
				continue
			}
			r := []any{}
			if timeFirst {
				r = append(r, timestamp, fmt.Sprintf("%s%s", namePrefix, field.Name), field.Value.Raw())
			} else {
				r = append(r, fmt.Sprintf("%s%s", namePrefix, field.Name), timestamp, field.Value.Raw())
			}
			rs = append(rs, r)
		}
	}
	jw.enc.Encode(rs)

	return nil
}
