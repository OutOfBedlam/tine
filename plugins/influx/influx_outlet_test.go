package influx_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/influx"
	_ "github.com/OutOfBedlam/tine/plugins/psutil"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type influxdbContainer struct {
	testcontainers.Container
	host_port string
}

func setupInfluxdb(ctx context.Context) (*influxdbContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "influxdb:1.8.10",
		ExposedPorts: []string{"8086/tcp"},
		Env: map[string]string{
			"INFLUXDB_DB":                "metrics",
			"INFLUXDB_HTTP_AUTH_ENABLED": "false",
		},
		WaitingFor: wait.ForLog("Listening on HTTP"),
	}
	influxC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	ip, err := influxC.Host(ctx)
	if err != nil {
		return nil, err
	}
	mappedPort, err := influxC.MappedPort(ctx, "8086")
	if err != nil {
		return nil, err
	}
	host_port := ip + ":" + mappedPort.Port()
	return &influxdbContainer{Container: influxC, host_port: host_port}, nil
}

func TestInfluxOutlet(t *testing.T) {
	ctx := context.Background()
	influxC, err := setupInfluxdb(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("skipping test on non-docker system")
		}
		fmt.Printf("Error: %v %T\n", err, err)
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := influxC.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	})

	dsl := fmt.Sprintf(`
		[log]
			path = "-"
			level = "debug"
			no_color = true
		[defaults]
			interval = "3s"
			count = 3
		[[inlets.load]]
			loads = [1,5,15]
		[[inlets.mem]]
		[[inlets.host]]
		[[flows.merge]]
			wait_limit = "5s"
			name_infix = "."
		[[outlets.influx]]
			## If the database does not exist, create first by running:
			## curl -XPOST 'http://localhost:8086/query' --data-urlencode 'q=CREATE DATABASE "metrics"'
			db = "metrics"
			path = "http://%s/write?db=metrics"
			tags = [
				{name="dc", value="us-east-1"},
				{name="env", value="prod"},
				{name="_in"}
			]
			## Write timeout, especially for the HTTP request
			timeout = "3s"
			## Debug mode for logging the response message from the InfluxDB 
			debug = true
	`, influxC.host_port)
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		t.Fatal(err)
	}

	if err := pipeline.Run(); err != nil {
		t.Fatal(err)
	}
	pipeline.Stop()
}
