package json

import (
	gojson "encoding/json"
	"fmt"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterEncoder(&engine.EncoderReg{
		Name:        "json",
		Factory:     NewJSONEncoder,
		ContentType: "application/x-ndjson",
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
	}

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
			} else if f.Type() == engine.BINARY {
				r[f.Name] = "[BINARY]"
			} else if f.Type() == engine.FLOAT {
				if jw.FormatOption.Decimal == 0 {
					r[f.Name] = int(f.Value.Raw().(float64))
				} else if jw.FormatOption.Decimal > 0 {
					r[f.Name] = JsonFloat{Value: f.Value.Raw().(float64), Decimal: jw.FormatOption.Decimal}
				} else {
					r[f.Name] = f.Value.Raw()
				}
			} else {
				r[f.Name] = f.Value.Raw()
			}
		}
		jw.enc.Encode(r)
	}
	return nil
}

type JsonFloat struct {
	Value   float64
	Decimal int
}

func (l JsonFloat) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%.*f", l.Decimal, l.Value)
	return []byte(s), nil
}
