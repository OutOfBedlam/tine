package http

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "http",
		Factory: HttpOutlet,
	})
}

func HttpOutlet(ctx *engine.Context) engine.Outlet {
	return &httpOutlet{
		ctx: ctx,
	}
}

type httpOutlet struct {
	ctx    *engine.Context
	client *http.Client

	addr        string
	method      string
	successCode int

	writerConf engine.Config
}

func (ho *httpOutlet) Open() error {
	conf := ho.ctx.Config()
	ho.addr = conf.GetString("address", "")
	ho.method = conf.GetString("method", "POST")
	ho.successCode = conf.GetInt("success", 200)
	timeout := conf.GetDuration("timeout", 3*time.Second)

	ho.writerConf = conf.GetConfig("writer", engine.Config{"format": "csv"})

	// outlet common params
	slog.Debug("outlet.http", "address", ho.addr, "method", ho.method, "writer", ho.writerConf)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	ho.client = &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return nil
}

func (ho *httpOutlet) Close() error {
	return nil
}

func (ho *httpOutlet) Handle(recs []engine.Record) error {
	data := &bytes.Buffer{}
	w, err := engine.NewWriter(
		engine.WithWriter(data),
		engine.WithWriterConfig(ho.writerConf),
	)
	if err != nil {
		return err
	}
	w.Write(recs)

	var rsp *http.Response
	req, err := http.NewRequest(ho.method, ho.addr, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.ContentType)
	if w.ContentEncoding != "" {
		req.Header.Set("Content-Encoding", w.ContentEncoding)
	}
	rsp, err = ho.client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	w.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	if rsp.StatusCode != ho.successCode {
		slog.Warn("outlet.http", "status", rsp.Status, "response", string(body))
		return nil
	} else {
		slog.Debug("outlet.http", "status", rsp.Status, "response", string(body))
	}
	return nil
}
