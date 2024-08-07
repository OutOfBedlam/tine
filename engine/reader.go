package engine

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Reader struct {
	Format     string
	Timeformat string
	Timezone   string
	Compress   string
	Fields     []string
	Types      []string

	decoder Decoder
	raw     io.ReadCloser
}

type ReaderOption func(*Reader)

func NewReader(r io.Reader, cfg Config) (*Reader, error) {
	ret := &Reader{
		Format:     cfg.GetString("format", "csv"),
		Timeformat: cfg.GetString("timeformat", "s"),
		Timezone:   cfg.GetString("tz", "Local"),
		Compress:   cfg.GetString("compress", ""),
		Fields:     cfg.GetStringSlice("fields", []string{}),
		Types:      cfg.GetStringSlice("types", []string{}),
	}

	reg := GetDecoder(ret.Format)
	if reg == nil {
		return nil, fmt.Errorf("format %q not found", ret.Format)
	}
	timeformatter := &Timeformatter{
		format: ret.Timeformat,
	}
	if loc, err := time.LoadLocation(ret.Timezone); err == nil {
		timeformatter.loc = loc
	} else {
		timeformatter.loc = time.Local
	}

	if r == nil {
		ret.raw = io.NopCloser(os.Stdin)
	} else if rc, ok := r.(io.ReadCloser); ok {
		ret.raw = rc
	} else {
		ret.raw = io.NopCloser(r)
	}

	var types []Type
	if len(ret.Types) == 0 {
		types = make([]Type, len(ret.Fields))
		for i := range ret.Fields {
			types[i] = STRING
		}
	} else {
		if len(ret.Fields) != len(ret.Types) {
			return nil, fmt.Errorf("length of fields(%d) and types(%d) is not equal", len(ret.Fields), len(ret.Types))
		}
		types = make([]Type, len(ret.Types))
		for i := range ret.Types {
			switch strings.ToLower(ret.Types[i]) {
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
				// any types, do not use this types in other places
				types[i] = Type('?')
			default:
				return nil, fmt.Errorf("unknown type %q", ret.Types[i])
			}
		}
	}
	ret.decoder = reg.Factory(DecoderConfig{
		Reader:       ret.raw,
		Fields:       ret.Fields,
		Types:        types,
		FormatOption: ValueFormat{Timeformat: timeformatter},
	})

	compress := GetDecompressor(ret.Compress)
	if compress != nil {
		ret.raw = compress.Factory(ret.raw)
	}

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
