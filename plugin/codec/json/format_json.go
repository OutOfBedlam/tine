package json

import (
	gojson "encoding/json"
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
	enc *gojson.Encoder
}

func (jw *JSONEncoder) Encode(recs []engine.Record) error {
	if jw.enc == nil {
		jw.enc = gojson.NewEncoder(jw.Writer)
		jw.enc.SetIndent(jw.Prefix, jw.Indent)
	}

	rs := []map[string]any{}
	for _, rec := range recs {
		if rec.Empty() {
			continue
		}
		var ts time.Time = engine.Now()
		if v := rec.Tags().Get(engine.TAG_TIMESTAMP); v != nil {
			if t, ok := v.Time(); ok {
				ts = t
			}
		}
		var timestamp any
		if jw.FormatOption.Timeformat.IsEpoch() {
			timestamp = jw.FormatOption.Timeformat.Epoch(ts)
		} else {
			timestamp = jw.FormatOption.Timeformat.Format(ts)
		}
		r := map[string]any{engine.TAG_TIMESTAMP: timestamp}
		if v := rec.Tags().Get(engine.TAG_INLET); v != nil {
			if s, ok := v.String(); ok {
				r[engine.TAG_INLET] = s
			}
		}
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
