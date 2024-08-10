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
	engine.RegisterDecoder(&engine.DecoderReg{
		Name:    "json",
		Factory: NewJSONDecoder,
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
		if err := jw.enc.Encode(r); err != nil {
			// e.g. when http client socket closed, it will return error
			return err
		}
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

func NewJSONDecoder(c engine.DecoderConfig) engine.Decoder {
	return &JSONDecoder{
		DecoderConfig: c,
	}
}

type JSONDecoder struct {
	engine.DecoderConfig
	dec *gojson.Decoder
}

func (jd *JSONDecoder) Decode() ([]engine.Record, error) {
	if jd.dec == nil {
		jd.dec = gojson.NewDecoder(jd.Reader)
	}

	var ret = []engine.Record{}
	var retErr error
	for {
		var r map[string]any
		if err := jd.dec.Decode(&r); err != nil {
			retErr = err
			break
		}
		if rec, err := map2Record(r); err != nil {
			retErr = err
			break
		} else {
			ret = append(ret, rec)
		}
	}
	return ret, retErr
}

func map2Record(m map[string]any) (engine.Record, error) {
	rec := engine.NewRecord()
	for k, val := range m {
		switch v := val.(type) {
		case string:
			rec = rec.Append(engine.NewField(k, v))
		case float64:
			rec = rec.Append(engine.NewField(k, v))
		case bool:
			rec = rec.Append(engine.NewField(k, v))
		default:
			return nil, fmt.Errorf("unsupported type %T", v)
		}
	}
	return rec, nil
}
