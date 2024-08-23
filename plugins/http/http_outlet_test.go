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

var httpOutletServer *http.Server

func runOutletTestServer() (string, error) {
	httpOutletServer = &http.Server{
		Addr: "127.0.0.1:0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := &strings.Builder{}
			io.Copy(b, r.Body)
			fmt.Println(b.String())
			w.WriteHeader(200)
		}),
	}

	lsnr, err := net.Listen("tcp", httpOutletServer.Addr)
	if err != nil {
		return "", err
	}

	go http.Serve(lsnr, httpOutletServer.Handler)

	// Get the address the server is listening on
	addr := lsnr.Addr().String()
	return addr, nil
}

func ExampleHttpOutlet() {
	addr, err := runOutletTestServer()
	if err != nil {
		panic(err)
	}
	defer httpOutletServer.Close()

	dsl := fmt.Sprintf(`
	[[inlets.file]]
		data = [
			"a,1", 
			"b,2", 
			"c,3",
		]
		format = "csv"
	[[flows.select]]
		includes = ["**"]
	[[outlets.http]]
	    address = "http://%s"
		format = "json"
	`, addr)
	// Make the output time deterministic. so we can compare it.
	// This line is not needed in production code.
	serial := int64(0)
	engine.Now = func() time.Time { serial++; return time.Unix(1721954797+serial, 0) }
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
	// {"0":"a","1":"1","_in":"file","_ts":1721954798}
	// {"0":"b","1":"2","_in":"file","_ts":1721954799}
	// {"0":"c","1":"3","_in":"file","_ts":1721954800}
}
