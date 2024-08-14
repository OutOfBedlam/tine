package http_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/http"
)

var httpServer *http.Server

func runTestServer() (string, error) {
	httpServer = &http.Server{
		Addr: "127.0.0.1:0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"a":1, "b":{"c":true, "d":3.14}}` + "\n"))
			w.WriteHeader(200)
		}),
	}

	lsnr, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return "", err
	}

	go http.Serve(lsnr, httpServer.Handler)

	// Get the address the server is listening on
	addr := lsnr.Addr().String()
	return addr, nil
}

func ExampleHttpInlet() {
	addr, err := runTestServer()
	if err != nil {
		panic(err)
	}
	defer httpServer.Close()

	dsl := fmt.Sprintf(`
	[[inlets.http]]
		address = "http://%s"
		success = 200
		timeout = "3s"
		count = 1
	[[flows.select]]
		includes = ["**"]
	[[outlets.file]]
		format = "json"
	`, addr)
	// Make the output time deterministic. so we can compare it.
	// This line is not needed in production code.
	engine.Now = func() time.Time { return time.Unix(1721954797, 0) }
	// Create a new engine.
	out := &bytes.Buffer{}
	pipeline, err := engine.New(engine.WithConfig(dsl), engine.WithWriter(out))
	if err != nil {
		panic(err)
	}
	// Run the engine.
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	result, _ := io.ReadAll(out)
	fmt.Println(strings.TrimSpace(string(result)))
	// Output:
	// {"_in":"http","_ts":1721954797,"a":1,"b.c":true,"b.d":3.14}
}
