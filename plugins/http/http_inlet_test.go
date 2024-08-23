package http_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/http"
	"github.com/stretchr/testify/require"
)

var httpInletServer *http.Server

func runInletTestServer() (string, error) {
	httpInletServer = &http.Server{
		Addr: "127.0.0.1:0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"a":1, "b":{"c":true, "d":3.14, "str":"text"}, "arr":["first", 2, 3.14, true]}` + "\n"))
				w.WriteHeader(200)
			case "/binary":
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write([]byte{0x01, 0x02, 0x03})
				w.WriteHeader(200)
			default:
				w.WriteHeader(404)
			}
		}),
	}

	lsnr, err := net.Listen("tcp", httpInletServer.Addr)
	if err != nil {
		return "", err
	}

	go http.Serve(lsnr, httpInletServer.Handler)

	// Get the address the server is listening on
	addr := lsnr.Addr().String()
	return addr, nil
}

func TestHttpInlet(t *testing.T) {
	addr, err := runInletTestServer()
	if err != nil {
		panic(err)
	}
	defer httpInletServer.Close()

	tests := []struct {
		name       string
		path       string
		expectBody string
	}{
		{
			name:       "success",
			path:       "/",
			expectBody: `{"_in":"http","_ts":1721954797,"a":1,"arr[0]":"first","arr[1]":2,"arr[2]":3.14,"arr[3]":true,"b.c":true,"b.d":3.14,"b.str":"text"}`,
		},
		{
			name:       "not found",
			path:       "/notfound",
			expectBody: "",
		},
		{
			name:       "binary",
			path:       "/binary",
			expectBody: "",
		},
	}

	for _, tt := range tests {
		dsl := fmt.Sprintf(`
			[[inlets.http]]
				address = "http://%s%s"
				success = 200
				timeout = "3s"
				count = 1
			[[flows.select]]
				includes = ["**"]
			[[outlets.file]]
				format = "json"
			`, addr, tt.path)
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
		require.Equal(t, tt.expectBody, strings.TrimSpace(string(result)), tt.name)
	}
}
