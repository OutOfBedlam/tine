package engine

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type Reader struct {
	format     string
	timeformat string
	timezone   string
	compress   string
	fields     []string
	types      []string

	decoder Decoder
	raw     io.ReadCloser
}

type ReaderOption func(*Reader)

func WithReader(r io.Reader) ReaderOption {
	return func(rd *Reader) {
		rd.raw = io.NopCloser(r)
	}
}

func WithReaderConfig(cfg Config) ReaderOption {
	return func(rd *Reader) {
		rd.format = cfg.GetString("format", "csv")
		rd.timeformat = cfg.GetString("timeformat", "s")
		rd.timezone = cfg.GetString("tz", "Local")
		rd.compress = cfg.GetString("compress", "")
		rd.fields = cfg.GetStringSlice("fields", nil)
		rd.types = cfg.GetStringSlice("types", nil)
	}
}

func NewReader(opts ...ReaderOption) (*Reader, error) {
	ret := &Reader{}
	for _, opt := range opts {
		opt(ret)
	}

	reg := GetDecoder(ret.format)
	if reg == nil {
		return nil, fmt.Errorf("format %q not found", ret.format)
	}
	timeformatter := &Timeformatter{
		format: ret.timeformat,
	}
	if loc, err := time.LoadLocation(ret.timezone); err == nil {
		timeformatter.loc = loc
	} else {
		timeformatter.loc = time.Local
	}

	if ret.raw == nil {
		return nil, fmt.Errorf("no reader specified")
	}

	var types []Type
	if len(ret.types) == 0 {
		types = make([]Type, len(ret.fields))
		for i := range ret.fields {
			types[i] = STRING
		}
	} else {
		if len(ret.fields) != len(ret.types) {
			return nil, fmt.Errorf("length of fields(%d) and types(%d) is not equal", len(ret.fields), len(ret.types))
		}
		types = make([]Type, len(ret.types))
		for i := range ret.types {
			switch strings.ToLower(ret.types[i]) {
			case "string":
				types[i] = STRING
			case "int":
				types[i] = INT
			case "uint":
				types[i] = UINT
			case "float":
				types[i] = FLOAT
			case "time":
				types[i] = TIME
			case "bool", "boolean":
				types[i] = BOOL
			case "any":
				// any types, do not use this typs in other places
				types[i] = Type('?')
			default:
				return nil, fmt.Errorf("unknown type %q", ret.types[i])
			}
		}
	}
	compress := GetDecompressor(ret.compress)
	if compress != nil {
		ret.raw = compress.Factory(ret.raw)
	}

	ret.decoder = reg.Factory(DecoderConfig{
		Reader:        ret.raw,
		Timeformatter: timeformatter,
		Fields:        ret.fields,
		Types:         types,
	})
	return ret, nil
}

func (rd *Reader) Read() ([]Record, error) {
	return rd.decoder.Decode()
}

func (rd *Reader) Close() error {
	if rd.raw != nil {
		return rd.raw.Close()
	}
	return nil
}
