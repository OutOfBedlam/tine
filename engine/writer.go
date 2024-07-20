package engine

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Writer struct {
	ContentType     string
	ContentEncoding string

	format       string
	subFormat    string
	outputIndent string
	outputPrefix string
	timeformat   string
	timezone     string
	decimal      int
	compress     string
	fields       []string

	encoder Encoder
	raw     io.WriteCloser
}

type WriterOptions func(*Writer)

func WithWriterConfig(cfg Config) WriterOptions {
	return func(e *Writer) {
		e.format = cfg.GetString("format", e.format)
		e.subFormat = cfg.GetString("subformat", e.subFormat)
		e.outputIndent = cfg.GetString("indent", e.outputIndent)
		e.outputPrefix = cfg.GetString("prefix", e.outputPrefix)
		e.timeformat = cfg.GetString("timeformat", e.timeformat)
		e.timezone = cfg.GetString("tz", e.timezone)
		e.decimal = cfg.GetInt("decimal", e.decimal)
		e.compress = cfg.GetString("compress", e.compress)
		e.fields = cfg.GetStringArray("fields", e.fields)
	}
}

func NewWriter(w io.Writer, opts ...WriterOptions) (*Writer, error) {
	ret := &Writer{
		format:     "csv",
		decimal:    -1,
		timeformat: "s",
		timezone:   "Local",
		fields:     []string{},
		raw:        &NopCloser{w},
	}

	for _, opt := range opts {
		opt(ret)
	}

	reg := GetEncoder(ret.format)
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
		ret.raw = &NopCloser{io.Writer(os.Stdout)}
	}

	comppress := GetCompressor(ret.compress)
	if comppress != nil {
		ret.raw = comppress.Factory(ret.raw)
		ret.ContentEncoding = comppress.ContentEncoding
	} else {
		ret.ContentEncoding = ""
	}

	ret.encoder = reg.Factory(EncoderConfig{
		Writer:        ret.raw,
		Subformat:     ret.subFormat,
		Indent:        ret.outputIndent,
		Prefix:        ret.outputPrefix,
		Timeformatter: timeformatter,
		Fields:        ret.fields,
		Decimal:       ret.decimal,
	})
	ret.ContentType = reg.ContentType
	if ret.ContentType == "" {
		ret.ContentType = "applicaton/octet-stream"
	}
	return ret, nil
}

func (rw *Writer) Write(recs []Record) error {
	return rw.encoder.Encode(recs)
}

func (rw *Writer) Close() error {
	if rw.raw != nil {
		return rw.raw.Close()
	}
	return nil
}

type NopCloser struct {
	io.Writer
}

func (nc *NopCloser) Close() error {
	return nil
}
