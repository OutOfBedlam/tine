package engine

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Writer struct {
	Format       string
	Subformat    string
	OutputIndent string
	OutputPrefix string
	Timeformat   string
	Timezone     string
	Decimal      int
	Compress     string
	Fields       []string

	ContentType     string
	ContentEncoding string
	encoder         Encoder
	raw             io.WriteCloser
}

func NewWriter(w io.Writer, cfg Config) (*Writer, error) {
	ret := &Writer{
		Format:       cfg.GetString("format", "csv"),
		Subformat:    cfg.GetString("subformat", ""),
		OutputIndent: cfg.GetString("indent", ""),
		OutputPrefix: cfg.GetString("prefix", ""),
		Decimal:      cfg.GetInt("decimal", -1),
		Timeformat:   cfg.GetString("timeformat", "s"),
		Timezone:     cfg.GetString("tz", "Local"),
		Compress:     cfg.GetString("compress", ""),
		Fields:       cfg.GetStringArray("fields", []string{}),
		raw:          NopCloser(w),
	}

	reg := GetEncoder(ret.Format)
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

	if ret.raw == nil {
		ret.raw = NopCloser(os.Stdout)
	}

	comppress := GetCompressor(ret.Compress)
	if comppress != nil {
		ret.raw = comppress.Factory(ret.raw)
		ret.ContentEncoding = comppress.ContentEncoding
	} else {
		ret.ContentEncoding = ""
	}

	ret.encoder = reg.Factory(EncoderConfig{
		Writer:       ret.raw,
		Subformat:    ret.Subformat,
		Indent:       ret.OutputIndent,
		Prefix:       ret.OutputPrefix,
		Fields:       ret.Fields,
		FormatOption: ValueFormat{Timeformat: timeformatter, Decimal: ret.Decimal},
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

func NopCloser(w io.Writer) io.WriteCloser {
	return &nopCloser{w}
}

type nopCloser struct {
	io.Writer
}

func (nc *nopCloser) Close() error {
	return nil
}
