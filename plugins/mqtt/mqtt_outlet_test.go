package mqtt_test

import (
	"testing"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/mqtt"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/stretchr/testify/require"
)

var server *mqtt.Server

func TestMain(m *testing.M) {
	server = mqtt.New(&mqtt.Options{
		InlineClient: true,
	})
	_ = server.AddHook(new(auth.AllowHook), nil)
	tcp := listeners.NewTCP(listeners.Config{ID: "tcp", Address: "127.0.0.1:1883"})
	err := server.AddListener(tcp)
	if err != nil {
		panic(err)
	}
	go server.Serve()

	m.Run()

	server.Close()
}

func TestMqttOutlet(t *testing.T) {
	expect := `{"name":"a","rec.value":100}
{"name":"b","rec.value":200}
{"name":"c","rec.value":300}
`
	err := server.Subscribe("test/#", 1, func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		require.Equal(t, "test/json", pk.TopicName)
		require.Equal(t, expect, string(pk.Payload))
	})
	if err != nil {
		t.Fatal(err)
	}

	dsl := `
	[log]
		path = "-"
		level = "info"
		no_color = true
	[[inlets.file]]
		data = [
			"a,100",
			"b,200",
			"c,300",
		]
		format = "csv"
		fields = ["name", "rec.value"]
		types  = ["string", "int"]
	[[outlets.mqtt]]
		server = "tcp://127.0.0.1:1883"
		username = "user"
		password = "pass"
		topic = "test/json"
		qos = 1
		timeout = "3s"
		format = "json"
	`
	p, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		t.Fatal(err)
	}
	err = p.Run()
	if err != nil {
		t.Fatal(err)
	}
}
