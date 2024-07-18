package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

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
	enc *csv.Writer
}

func (cw *CSVEncoder) Encode(recs []engine.Record) error {
	if cw.enc == nil {
		cw.enc = csv.NewWriter(cw.Writer)
	}

	switch strings.ToLower(cw.Subformat) {
	default:
		return cw.encodeDefault(recs)
	case "name_time_value":
		return cw.encodeNameTimeValue(recs, false)
	case "time_name_value":
		return cw.encodeNameTimeValue(recs, true)
	}
}

func (cw *CSVEncoder) encodeDefault(recs []engine.Record) error {
	values := make([]string, 0, 16)
	for _, rec := range recs {
		for _, field := range rec.Fields(cw.Fields...) {
			if field == nil {
				values = append(values, "")
				continue
			}
			if field.Type == engine.BINARY {
				values = append(values, "[BINARY]")
				continue
			}
			values = append(values, field.StringWithFormat(cw.Timeformatter, cw.Decimal))
		}
		cw.enc.Write(values)
		values = values[:0]
	}
	cw.enc.Flush()
	return nil
}

func (cw *CSVEncoder) encodeNameTimeValue(recs []engine.Record, timeFirst bool) error {
	for _, rec := range recs {
		var ts time.Time
		if tsf := rec.Field(engine.FIELD_TIMESTAMP); tsf == nil {
			ts = time.Now()
		} else {
			ts = tsf.Value.(time.Time)
		}
		var namePrefix string
		if inf := rec.Field(engine.FIELD_INLET); inf == nil {
			namePrefix = ""
		} else {
			namePrefix = inf.Value.(string) + "."
		}
		for _, field := range rec.Fields(cw.Fields...) {
			if field == nil || field.Name == engine.FIELD_TIMESTAMP || field.Name == engine.FIELD_INLET {
				continue
			}
			if field.Type == engine.BINARY {
				continue
			}
			if timeFirst {
				cw.enc.Write([]string{
					cw.Timeformatter.Format(ts),
					fmt.Sprintf("%s%s", namePrefix, field.Name),
					field.StringWithFormat(cw.Timeformatter, cw.Decimal),
				})
			} else {
				cw.enc.Write([]string{
					fmt.Sprintf("%s%s", namePrefix, field.Name),
					cw.Timeformatter.Format(ts),
					field.StringWithFormat(cw.Timeformatter, cw.Decimal),
				})
			}
		}
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
		strVal := f.StringWithFormat(tf, decimal)
		ret = append(ret, f.Name, strVal)
	}
	return ret
}

func NewCSVDecoder(conf engine.DecoderConfig) engine.Decoder {
	return &CSVDecoder{
		raw:    conf.Reader,
		tf:     conf.Timeformatter,
		fields: conf.Fields,
		types:  conf.Types,
	}
}

type CSVDecoder struct {
	raw    io.Reader
	fields []string
	types  []engine.Type
	tf     *engine.Timeformatter

	dec *csv.Reader
}

func (cr *CSVDecoder) Decode() ([]engine.Record, error) {
	if cr.dec == nil {
		cr.dec = csv.NewReader(cr.raw)
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
		if len(cr.fields) == 0 {
			// there is no fields specified, accept all fields
			cols := []*engine.Field{}
			for idx, str := range fields {
				cols = append(cols, engine.NewStringField(strconv.Itoa(idx), str))
			}
			rec := engine.NewRecord(cols...)
			ret = append(ret, rec)
		} else {
			// input fields are specified
			cols := []*engine.Field{}
			for idx, str := range fields {
				var fieldName string
				if idx >= len(cr.fields) {
					break
				} else {
					fieldName = strings.ToUpper(cr.fields[idx])
				}
				if idx >= len(cr.types) {
					cols = append(cols, engine.NewStringField(fieldName, str))
				} else {
					switch cr.types[idx] {
					default:
						cols = append(cols, engine.NewStringField(fieldName, str))
					case engine.Type('?'):
						if strings.ToLower(str) == "true" {
							cols = append(cols, engine.NewBoolField(fieldName, true))
						} else if strings.ToLower(str) == "false" {
							cols = append(cols, engine.NewBoolField(fieldName, false))
						} else if f, err := strconv.ParseFloat(str, 64); err == nil {
							cols = append(cols, engine.NewFloatField(fieldName, f))
						} else if tm, err := cr.tf.Parse(str); err == nil {
							cols = append(cols, engine.NewTimeField(fieldName, tm))
						} else {
							cols = append(cols, engine.NewStringField(fieldName, str))
						}
					case engine.STRING:
						cols = append(cols, engine.NewStringField(fieldName, str))
					case engine.INT:
						if i, err := strconv.ParseInt(str, 10, 64); err != nil {
							cols = append(cols, engine.NewStringField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewIntField(fieldName, i))
						}
					case engine.UINT:
						if i, err := strconv.ParseInt(str, 10, 64); err != nil {
							cols = append(cols, engine.NewStringField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewIntField(fieldName, i))
						}
					case engine.FLOAT:
						if f, err := strconv.ParseFloat(str, 64); err != nil {
							cols = append(cols, engine.NewStringField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewFloatField(fieldName, f))
						}
					case engine.BOOL:
						if b, err := strconv.ParseBool(str); err != nil {
							cols = append(cols, engine.NewStringField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewBoolField(fieldName, b))
						}
					case engine.TIME:
						if tm, err := cr.tf.Parse(str); err != nil {
							cols = append(cols, engine.NewStringField(fieldName, str+"; "+err.Error()))
						} else {
							cols = append(cols, engine.NewTimeField(fieldName, tm))
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
