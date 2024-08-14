package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "http",
		Factory: HttpInlet,
	})
}

func HttpInlet(ctx *engine.Context) engine.Inlet {
	return &httpInlet{
		ctx: ctx,
	}
}

type httpInlet struct {
	ctx      *engine.Context
	client   *http.Client
	runCount int64

	addr          string
	successCode   int
	runCountLimit int64
}

var _ = engine.Inlet((*httpInlet)(nil))

func (hi *httpInlet) Open() error {
	hi.addr = hi.ctx.Config().GetString("address", "")
	hi.successCode = hi.ctx.Config().GetInt("success", 200)
	timeout := hi.ctx.Config().GetDuration("timeout", 3*time.Second)
	hi.runCountLimit = int64(hi.ctx.Config().GetInt("count", 1))

	slog.Debug("inlet.http", "address", hi.addr, "success", hi.successCode, "timeout", timeout)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	hi.client = &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return nil
}

func (hi *httpInlet) Close() error {
	return nil
}

func (hi *httpInlet) Interval() time.Duration {
	return hi.ctx.Config().GetDuration("interval", hi.client.Timeout)
}

func (hi *httpInlet) Process(next engine.InletNextFunc) {
	if hi.runCountLimit > 0 && atomic.LoadInt64(&hi.runCount) >= hi.runCountLimit {
		next(nil, io.EOF)
		return
	}
	runCount := atomic.AddInt64(&hi.runCount, 1)

	rsp, err := hi.client.Get(hi.addr)
	if err != nil {
		next(nil, err)
		return
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		next(nil, err)
		return
	}

	if rsp.StatusCode != hi.successCode {
		slog.Warn("inlet.http", "status", rsp.StatusCode, "body", string(body))
		next(nil, nil)
		return
	}

	var resultErr error
	if hi.runCountLimit > 0 && runCount > hi.runCountLimit {
		resultErr = io.EOF
	}

	// TODO: support other content-type
	// [x] application/json
	// [ ] application/x-ndjson
	// [ ] text/csv
	if contentType := rsp.Header.Get("Content-Type"); strings.Contains(contentType, "application/json") {
		obj := map[string]any{}
		if err := json.Unmarshal(body, &obj); err != nil {
			slog.Warn("inlet.http", "status", rsp.StatusCode, "unmarshal error", err.Error())
			next(nil, err)
			return
		}
		ret := Map2Record("", obj)
		next([]engine.Record{ret}, resultErr)
		return
	} else {
		slog.Warn("inlet.http", "status", rsp.StatusCode, "unsupported content-type", contentType)
		next(nil, resultErr)
		return
	}
}

func Map2Record(prefix string, obj map[string]any) engine.Record {
	ret := engine.NewRecord()
	for k, v := range obj {
		switch v := v.(type) {
		case float64:
			ret = ret.Append(engine.NewField(prefix+k, v))
		case int:
			ret = ret.Append(engine.NewField(prefix+k, int64(v)))
		case int64:
			ret = ret.Append(engine.NewField(prefix+k, v))
		case string:
			ret = ret.Append(engine.NewField(prefix+k, v))
		case bool:
			ret = ret.Append(engine.NewField(prefix+k, v))
		case time.Time:
			ret = ret.Append(engine.NewField(prefix+k, v))
		case []byte:
			ret = ret.Append(engine.NewField(prefix+k, v))
		case map[string]any:
			subRec := Map2Record(prefix+k+".", v)
			ret = ret.Append(subRec.Fields()...)
		case []any:
			// TODO: support array
			continue
		default:
			continue
		}
	}
	return ret
}
