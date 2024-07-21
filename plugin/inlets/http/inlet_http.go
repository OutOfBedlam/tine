package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
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
	ctx    *engine.Context
	client *http.Client

	addr        string
	successCode int
}

var _ = engine.PullInlet((*httpInlet)(nil))

func (hi *httpInlet) Open() error {
	hi.addr = hi.ctx.Config().GetString("address", "")
	hi.successCode = hi.ctx.Config().GetInt("success", 200)
	timeout := hi.ctx.Config().GetDuration("timeout", 3*time.Second)

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

func (hi *httpInlet) Pull() ([]engine.Record, error) {
	rsp, err := hi.client.Get(hi.addr)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != hi.successCode {
		slog.Warn("inlet.http", "status", rsp.StatusCode, "body", string(body))
		return nil, nil
	}

	if contentType := rsp.Header.Get("Content-Type"); strings.Contains(contentType, "application/json") {
		obj := map[string]any{}
		if err := json.Unmarshal(body, &obj); err != nil {
			slog.Warn("inlet.http", "status", rsp.StatusCode, "unmarshal error", err.Error())
			return nil, err
		}
		ret := Map2Records("http.", obj)
		return ret, nil
	} else {
		slog.Warn("inlet.http", "status", rsp.StatusCode, "unsupported content-type", contentType)
		return nil, nil
	}
}

func Map2Records(prefix string, obj map[string]any) []engine.Record {
	ret := []engine.Record{}
	for k, v := range obj {
		r := engine.NewRecord()
		subrecs := []engine.Record{}
		switch v := v.(type) {
		case float64:
			r = r.Append(engine.NewFloatField(prefix+k, v))
		case int:
			r = r.Append(engine.NewIntField(prefix+k, int64(v)))
		case int64:
			r = r.Append(engine.NewIntField(prefix+k, v))
		case string:
			r = r.Append(engine.NewStringField(prefix+k, v))
		case bool:
			r = r.Append(engine.NewBoolField(prefix+k, v))
		case time.Time:
			r = r.Append(engine.NewTimeField(prefix+k, v))
		case []byte:
			r = r.Append(engine.NewBinaryField(prefix+k, engine.NewBinaryValue(v)))
		case map[string]any:
			subrecs = append(subrecs, Map2Records(prefix+k+".", v)...)
		case []any:
			// TODO: support array
			continue
		default:
			continue
		}
		ret = append(ret, r)
		ret = append(ret, subrecs...)
	}
	return ret
}
