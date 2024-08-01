package json

import (
	gojson "encoding/json"

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
