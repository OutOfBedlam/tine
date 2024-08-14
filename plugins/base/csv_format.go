package base

import (
	gocsv "encoding/csv"
	"strconv"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterEncoder(&engine.EncoderReg{
		Name:        "csv",
		Factory:     NewCSVEncoder,
		ContentType: "text/csv",
	})
	engine.RegisterDecoder(&engine.DecoderReg{
		Name:    "csv",
		Factory: NewCSVDecoder,
	})
}

func NewCSVEncoder(c engine.EncoderConfig) engine.Encoder {
	return &CSVEncoder{EncoderConfig: c}
}

type CSVEncoder struct {
	engine.EncoderConfig
	enc *gocsv.Writer
}

func (cw *CSVEncoder) Encode(recs []engine.Record) error {
	if cw.enc == nil {
		cw.enc = gocsv.NewWriter(cw.Writer)
	}

	return cw.encodeDefault(recs)
}

func (cw *CSVEncoder) encodeDefault(recs []engine.Record) error {
	values := make([]string, 0, 16)
	for _, rec := range recs {
		for _, field := range rec.Fields(cw.Fields...) {
			if field == nil {
				values = append(values, "")
				continue
			}
			if field.Type() == engine.BINARY {
				values = append(values, "[BINARY]")
				continue
			}
			values = append(values, field.Value.Format(cw.FormatOption))
		}
		if err := cw.enc.Write(values); err != nil {
			// e.g. when http client socket closed, it will return error
			return err
		}
		values = values[:0]
	}
	cw.enc.Flush()
	return nil
}

func StringFields(r engine.Record) []string {
	return StringFieldsWithFormat(r, engine.DefaultTimeformatter, -1)
}

func StringFieldsWithFormat(r engine.Record, tf *engine.Timeformatter, decimal int) []string {
	ret := []string{}
	for _, f := range r.Fields() {
		strVal := f.Value.Format(engine.ValueFormat{Timeformat: tf, Decimal: decimal})
		ret = append(ret, f.Name, strVal)
	}
	return ret
}

func NewCSVDecoder(conf engine.DecoderConfig) engine.Decoder {
	return &CSVDecoder{
		DecoderConfig: conf,
	}
}

type CSVDecoder struct {
	engine.DecoderConfig
	dec *gocsv.Reader
}

func (cr *CSVDecoder) Decode() ([]engine.Record, error) {
	if cr.dec == nil {
		cr.dec = gocsv.NewReader(cr.Reader)
	}
	var ret = []engine.Record{}
	var retErr error
	for {
		fields, err := cr.dec.Read()
		if err != nil {
			retErr = err
			break
		}
		if len(fields) == 0 {
			continue
		}
		if len(cr.Fields) == 0 {
			// there is no fields specified, accept all fields
			cols := []*engine.Field{}
			for idx, str := range fields {
				cols = append(cols, engine.NewField(strconv.Itoa(idx), str))
			}
			rec := engine.NewRecord(cols...)
			ret = append(ret, rec)
		} else {
			// input fields are specified
			cols := []*engine.Field{}
			for idx, str := range fields {
				var fieldName string
				if idx >= len(cr.Fields) {
					break
				} else {
					fieldName = cr.Fields[idx]
				}
				if idx >= len(cr.Types) {
					cols = append(cols, engine.NewField(fieldName, str))
				} else {
					switch cr.Types[idx] {
					default:
						cols = append(cols, engine.NewField(fieldName, str))
					case engine.Type('?'):
						if strings.ToLower(str) == "true" {
							cols = append(cols, engine.NewField(fieldName, true))
						} else if strings.ToLower(str) == "false" {
							cols = append(cols, engine.NewField(fieldName, false))
						} else if f, err := strconv.ParseFloat(str, 64); err == nil {
							cols = append(cols, engine.NewField(fieldName, f))
						} else if tm, err := cr.FormatOption.Timeformat.Parse(str); err == nil {
							cols = append(cols, engine.NewField(fieldName, tm))
						} else {
							cols = append(cols, engine.NewField(fieldName, str))
						}
					case engine.STRING:
						cols = append(cols, engine.NewField(fieldName, str))
					case engine.INT:
						if i, err := strconv.ParseInt(str, 10, 64); err != nil {
							cols = append(cols, engine.NewField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewField(fieldName, i))
						}
					case engine.UINT:
						if i, err := strconv.ParseInt(str, 10, 64); err != nil {
							cols = append(cols, engine.NewField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewField(fieldName, i))
						}
					case engine.FLOAT:
						if f, err := strconv.ParseFloat(str, 64); err != nil {
							cols = append(cols, engine.NewField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewField(fieldName, f))
						}
					case engine.BOOL:
						if b, err := strconv.ParseBool(str); err != nil {
							cols = append(cols, engine.NewField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewField(fieldName, b))
						}
					case engine.TIME:
						if tm, err := cr.FormatOption.Timeformat.Parse(str); err != nil {
							cols = append(cols, engine.NewField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewField(fieldName, tm))
						}
					}
				}
			}
			rec := engine.NewRecord(cols...)
			ret = append(ret, rec)
		}
	}
	return ret, retErr
}
