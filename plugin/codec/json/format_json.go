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
			if f.Type == engine.TIME {
				if jw.Timeformatter.IsEpoch() {
					r[f.Name] = jw.Timeformatter.Epoch(f.Value.(time.Time))
				} else {
					r[f.Name] = jw.Timeformatter.Format(f.Value.(time.Time))
				}
			} else {
				r[f.Name] = f.Value
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
			ts = tsf.Value.(time.Time)
		}
		var timestamp any
		if jw.Timeformatter.IsEpoch() {
			timestamp = jw.Timeformatter.Epoch(ts)
		} else {
			timestamp = jw.Timeformatter.Format(ts)
		}
		var namePrefix string
		if inf := rec.Field(engine.FIELD_INLET); inf == nil {
			namePrefix = ""
		} else {
			namePrefix = inf.Value.(string) + "."
		}
		for _, field := range rec.Fields(jw.Fields...) {
			if field == nil || field.Name == engine.FIELD_TIMESTAMP || field.Name == engine.FIELD_INLET {
				continue
			}

			r := []any{}
			if timeFirst {
				r = append(r, timestamp, fmt.Sprintf("%s%s", namePrefix, field.Name), field.Value)
			} else {
				r = append(r, fmt.Sprintf("%s%s", namePrefix, field.Name), timestamp, field.Value)
			}
			rs = append(rs, r)
		}
	}
	jw.enc.Encode(rs)

	return nil
}
