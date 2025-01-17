package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Writer struct {
	Format     string
	Timeformat string
	Timezone   string
	Decimal    int
	Compress   string
	Fields     []string

	ContentType     string
	ContentEncoding string
	encoder         Encoder
	raw             io.WriteCloser
}

func NewWriter(w io.Writer, cfg Config) (*Writer, error) {
	ret := &Writer{
		Format:     cfg.GetString("format", "csv"),
		Decimal:    cfg.GetInt("decimal", -1),
		Timeformat: cfg.GetString("timeformat", "s"),
		Timezone:   cfg.GetString("tz", "Local"),
		Compress:   cfg.GetString("compress", ""),
		Fields:     cfg.GetStringSlice("fields", []string{}),
		raw:        NopCloser(w),
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

	compress := GetCompressor(ret.Compress)
	if compress != nil {
		ret.raw = compress.Factory(ret.raw)
		ret.ContentEncoding = compress.ContentEncoding
	} else {
		ret.ContentEncoding = ""
	}

	ret.encoder = reg.Factory(EncoderConfig{
		Writer:       ret.raw,
		Fields:       ret.Fields,
		FormatOption: ValueFormat{Timeformat: timeformatter, Decimal: ret.Decimal},
	})
	ret.ContentType = reg.ContentType
	if ret.ContentType == "" {
		ret.ContentType = "application/octet-stream"
	}
	return ret, nil
}

func (rw *Writer) Write(recs []Record) error {
	var flusher Flusher
	if nc, ok := rw.raw.(*nopCloser); ok {
		if fr, ok := nc.Writer.(http.Flusher); ok {
			flusher = fr
		}
	}
	if err := rw.encoder.Encode(recs); err != nil {
		//fmt.Println("handle - error", err)
		return err
	} else {
		if flusher != nil {
			// flusher is important when it responds to http client
			// in 'Transfer-Encoding: chunked'
			flusher.Flush()
		}
		return nil
	}
}

func (rw *Writer) Close() error {
	if rw.raw != nil {
		return rw.raw.Close()
	}
	return nil
}

type Flusher interface {
	Flush()
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
