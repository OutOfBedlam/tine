package influx

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	ilp "github.com/influxdata/line-protocol"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "influx",
		Factory: InfluxOutlet,
	})
}

func InfluxOutlet(ctx *engine.Context) engine.Outlet {
	return &influxOutlet{ctx: ctx}
}

type influxOutlet struct {
	sync.RWMutex
	ctx            *engine.Context
	enc            *ilp.Encoder
	out            io.WriteCloser
	outBytes       *bytes.Buffer
	outClient      *httpWriter
	dbName         string
	tags           map[string]string
	tagsFromRecord []string
	totalWritten   uint64
}

var _ = engine.Outlet((*influxOutlet)(nil))

func (io *influxOutlet) Open() error {
	conf := io.ctx.Config()

	io.dbName = conf.GetString("db", "metrics")

	io.tags = make(map[string]string)
	for _, cfg := range conf.GetConfigSlice("tags", nil) {
		name := cfg.GetString("name", "")
		value := cfg.GetString("value", "")
		if name == "" {
			continue
		}
		if value == "" {
			io.tagsFromRecord = append(io.tagsFromRecord, name)
		} else {
			io.tags[name] = value
		}
	}

	path := conf.GetString("path", "")
	timeout := conf.GetDuration("timeout", 3*time.Second)
	debug := conf.GetBool("debug", false)
	switch path {
	case "":
		io.out = engine.NopCloser(io.ctx.Writer())
	case "-":
		io.out = engine.NopCloser(os.Stdout)
	default:
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			transport := &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			}
			client := &http.Client{
				Transport: transport,
				Timeout:   timeout,
			}
			hw := &httpWriter{
				client: client,
				addr:   path,
			}
			hw.debug = debug
			hw.log = func(lvl slog.Level, status int, body string, size int) {
				io.ctx.Log(lvl, "outlets.influx request", "status", status, "body", body, "sent", size)
			}
			io.outClient = hw
			io.outBytes = &bytes.Buffer{}
			io.out = engine.NopCloser(io.outBytes)
		} else {
			if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
				return err
			} else {
				io.out = f
			}
		}
	}
	io.enc = ilp.NewEncoder(io.out)
	return nil
}

func (io *influxOutlet) Close() error {
	if io.out != nil {
		io.out.Close()
	}
	if io.outClient != nil {
		io.outClient.client.CloseIdleConnections()
	}
	return nil
}

func (io *influxOutlet) Handle(recs []engine.Record) error {
	io.Lock()
	defer io.Unlock()
	for _, rec := range recs {
		m, _ := io.rec2Metric(rec)
		n, err := io.enc.Encode(m)
		if err != nil {
			return err
		}
		atomic.AddUint64(&io.totalWritten, uint64(n))
	}
	if io.outBytes != nil && io.outClient != nil {
		io.outClient.Write(io.outBytes.Bytes())
		io.outBytes.Reset()
	}
	return nil
}

func (io *influxOutlet) rec2Metric(rec engine.Record) (ilp.Metric, error) {
	tags := make(map[string]string) // Initialize the tags map
	for k, v := range io.tags {
		if v != "" { // do not add empty strings
			tags[k] = v
		}
	}

	fieldsTakenToTags := []string{}

	for _, k := range io.tagsFromRecord {
		var tagName = ""
		var tagValue *engine.Value
		if strings.HasPrefix(k, "#") {
			if v := rec.Tags().Get(k[1:]); v != nil {
				tagName = k[1:]
				tagValue = v
			}
		} else {
			if v := rec.Field(k); v != nil {
				tagName = k
				tagValue = v.Value
				fieldsTakenToTags = append(fieldsTakenToTags, k)
			}
		}
		if tagName == "" || tagValue == nil || tagValue.IsNull() {
			continue
		}
		switch tagValue.Type() {
		case engine.TIME:
			ts, _ := tagValue.Time()
			tags[tagName] = fmt.Sprintf("%d", ts.UnixNano())
		default:
			if s, ok := tagValue.String(); ok && s != "" {
				// do not add empty strings
				tags[tagName] = s
			}
		}
	}

	fields := make(map[string]interface{}) // Initialize the fields map
	for _, f := range rec.Fields() {
		if slices.Contains(fieldsTakenToTags, f.Name) {
			// excludes fields that were taken to tags
			continue
		}
		switch f.Type() {
		case engine.INT:
			if v, ok := f.Value.Int64(); ok {
				fields[f.Name] = v
			}
		case engine.UINT:
			if v, ok := f.Value.Uint64(); ok {
				fields[f.Name] = v
			}
		case engine.FLOAT:
			if v, ok := f.Value.Float64(); ok {
				fields[f.Name] = v
			}
		case engine.STRING:
			if v, ok := f.Value.String(); ok {
				if v != "" { // do not add empty strings
					fields[f.Name] = v
				}
			}
		case engine.BOOL:
			if v, ok := f.Value.Bool(); ok {
				fields[f.Name] = v
			}
		case engine.TIME:
			if v, ok := f.Value.Time(); ok {
				fields[f.Name] = v
			}
		}
	}
	var tm time.Time
	if v := rec.Tags().Get(engine.TAG_TIMESTAMP); v != nil {
		if ts, ok := v.Time(); ok {
			tm = ts
		}
	}
	if tm.IsZero() {
		tm = time.Now()
	}
	return ilp.New(io.dbName, tags, fields, tm)
}

type httpWriter struct {
	client *http.Client
	addr   string
	debug  bool
	log    func(lvl slog.Level, status int, body string, size int)
}

func (hw *httpWriter) Write(p []byte) (n int, err error) {
	req, err := http.NewRequest("POST", hw.addr, strings.NewReader(string(p)))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "text/plain")
	rsp, err := hw.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(rsp.Body)
		hw.log(slog.LevelError, rsp.StatusCode, string(body), len(p))
	} else if hw.debug {
		body, _ := io.ReadAll(rsp.Body)
		hw.log(slog.LevelDebug, rsp.StatusCode, string(body), len(p))
	}
	return len(p), nil
}
