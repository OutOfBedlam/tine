package nats

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	gonatsd "github.com/nats-io/nats-server/v2/server"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "nats",
		Factory: NatsInlet,
	})
}

func NatsInlet(ctx *engine.Context) engine.Inlet {
	server := ctx.Config().GetString("server", "")
	timeout := ctx.Config().GetDuration("timeout", 3*time.Second)

	return &natsInlet{
		ctx:     ctx,
		Server:  server,
		Timeout: timeout,
	}
}

type natsInlet struct {
	ctx     *engine.Context
	Server  string
	Timeout time.Duration

	address *url.URL
	client  *http.Client
}

var _ = engine.PullInlet((*natsInlet)(nil))

func (ni *natsInlet) Open() error {
	address, err := url.Parse(ni.Server)
	if err != nil {
		return err
	}
	address.Path = path.Join(address.Path, "varz")
	ni.address = address
	return nil
}

func (ni *natsInlet) Close() error {
	return nil
}

func (ni *natsInlet) Interval() time.Duration {
	return ni.ctx.Config().GetDuration("interval", ni.Timeout)
}

func (ni *natsInlet) Pull() ([]engine.Record, error) {
	if ni.client == nil {
		timeout := ni.Timeout
		if timeout == time.Duration(0) {
			timeout = 5 * time.Second
		}
		ni.client = createHttpClient(timeout)
	}
	resp, err := ni.client.Get(ni.address.String())
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	stats := new(gonatsd.Varz)
	err = json.Unmarshal(data, stats)
	if err != nil {
		return nil, err
	}

	return []engine.Record{
		engine.NewRecord(
			engine.NewStringField("server_id", stats.ID),
			engine.NewStringField("server_name", stats.Name),
			engine.NewStringField("version", stats.Version),
			engine.NewStringField("host", stats.Host),
			engine.NewIntField("port", int64(stats.Port)),
			engine.NewUintField("uptime", uint64((stats.Now.Sub(stats.Start)).Seconds())),
			engine.NewIntField("mem", stats.Mem),
			engine.NewIntField("cores", int64(stats.Cores)),
			engine.NewIntField("gomaxprocs", int64(stats.MaxProcs)),
			engine.NewFloatField("cpu", stats.CPU),
			engine.NewIntField("connections", int64(stats.Connections)),
			engine.NewUintField("total_connections", stats.TotalConnections),
			engine.NewIntField("in_msgs", stats.InMsgs),
			engine.NewIntField("out_msgs", stats.OutMsgs),
			engine.NewIntField("in_bytes", stats.InBytes),
			engine.NewIntField("out_bytes", stats.OutBytes),
			engine.NewIntField("slow_consumers", stats.SlowConsumers),
			engine.NewIntField("subscriptions", int64(stats.Subscriptions)),
		),
	}, nil
}

func createHttpClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
