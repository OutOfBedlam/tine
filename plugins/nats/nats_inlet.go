package nats

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	gonatsd "github.com/nats-io/nats-server/v2/server"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "nats_varz",
		Factory: NatsVarzInlet,
	})
}

func NatsVarzInlet(ctx *engine.Context) engine.Inlet {
	server := ctx.Config().GetString("server", "")
	timeout := ctx.Config().GetDuration("timeout", 3*time.Second)

	return &natsVarzInlet{
		ctx:     ctx,
		Server:  server,
		Timeout: timeout,
	}
}

type natsVarzInlet struct {
	ctx     *engine.Context
	Server  string
	Timeout time.Duration

	address *url.URL
	client  *http.Client
}

var _ = engine.Inlet((*natsVarzInlet)(nil))

func (ni *natsVarzInlet) Open() error {
	address, err := url.Parse(ni.Server)
	if err != nil {
		return err
	}
	ni.address = address
	return nil
}

func (ni *natsVarzInlet) Close() error {
	return nil
}

func (ni *natsVarzInlet) Interval() time.Duration {
	return ni.ctx.Config().GetDuration("interval", ni.Timeout)
}

func (ni *natsVarzInlet) Process(next engine.InletNextFunc) {
	if ni.client == nil {
		timeout := ni.Timeout
		if timeout == time.Duration(0) {
			timeout = 5 * time.Second
		}
		ni.client = createHttpClient(timeout)
	}
	resp, err := ni.client.Get(ni.address.String())
	if err != nil {
		next(nil, err)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		next(nil, err)
		return
	}
	resp.Body.Close()

	stats := new(gonatsd.Varz)
	err = json.Unmarshal(data, stats)
	if err != nil {
		next(nil, err)
		return
	}

	next([]engine.Record{
		engine.NewRecord(
			engine.NewField("server_id", stats.ID),
			engine.NewField("server_name", stats.Name),
			engine.NewField("version", stats.Version),
			engine.NewField("host", stats.Host),
			engine.NewField("port", int64(stats.Port)),
			engine.NewField("uptime", uint64((stats.Now.Sub(stats.Start)).Seconds())),
			engine.NewField("mem", stats.Mem),
			engine.NewField("cores", int64(stats.Cores)),
			engine.NewField("gomaxprocs", int64(stats.MaxProcs)),
			engine.NewField("cpu", stats.CPU),
			engine.NewField("connections", int64(stats.Connections)),
			engine.NewField("total_connections", stats.TotalConnections),
			engine.NewField("in_msgs", stats.InMsgs),
			engine.NewField("out_msgs", stats.OutMsgs),
			engine.NewField("in_bytes", stats.InBytes),
			engine.NewField("out_bytes", stats.OutBytes),
			engine.NewField("slow_consumers", stats.SlowConsumers),
			engine.NewField("subscriptions", int64(stats.Subscriptions)),
		),
	}, nil)
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
