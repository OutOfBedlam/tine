package mqtt

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	paho "github.com/eclipse/paho.mqtt.golang"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "mqtt",
		Factory: MqttOutlet,
	})
}

func MqttOutlet(ctx *engine.Context) engine.Outlet {
	return &mqttOutlet{ctx: ctx}
}

var serial uint64

type mqttOutlet struct {
	ctx *engine.Context

	host    string
	topic   string
	qos     byte
	timeout time.Duration
	client  paho.Client
}

func (mo *mqttOutlet) Open() error {
	atomic.AddUint64(&serial, 1)

	conf := mo.ctx.Config()
	mo.host = conf.GetString("server", mo.host)
	mo.topic = conf.GetString("topic", mo.topic)
	mo.qos = byte(conf.GetInt("qos", 1))
	mo.timeout = conf.GetDuration("timeout", 3*time.Second)

	opts := paho.NewClientOptions()
	opts.SetCleanSession(true)
	opts.SetConnectRetry(false)
	opts.SetAutoReconnect(true)
	opts.SetProtocolVersion(4)
	opts.SetClientID(fmt.Sprintf("mqtt-%d", serial))
	opts.AddBroker(mo.host)
	opts.SetKeepAlive(60 * time.Second)

	mo.client = paho.NewClient(opts)
	token := mo.client.Connect()
	token.WaitTimeout(mo.timeout)
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mo *mqttOutlet) Close() error {
	if mo.client != nil {
		mo.client.Disconnect(1000)
	}
	return nil
}

func (mo *mqttOutlet) Handle(recs []engine.Record) error {
	data := &bytes.Buffer{}
	w, err := engine.NewWriter(data, mo.ctx.Config())
	if err != nil {
		return err
	}
	w.Write(recs)

	tok := mo.client.Publish(mo.topic, mo.qos, false, data.Bytes())
	tok.WaitTimeout(mo.timeout)
	if tok.Error() != nil {
		slog.Error("outlet.mqtt", "error", tok.Error().Error())
	}
	return nil
}
