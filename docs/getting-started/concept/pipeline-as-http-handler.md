# Pipeline as Http Handler

A pipeline can serves as a http handler with Go standard `net/http`.

The example code below shows simplest web service that returns the CPU usage of the system.

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/all"
)

func main() {
	addr := "127.0.0.1:8080"
	router := http.NewServeMux()
	router.HandleFunc("GET /helloworld", engine.HttpHandleFunc(helloWorldPipeline))

	fmt.Printf("\nstart server http://%s\n\n", addr)
	http.ListenAndServe(addr, router)
}

const helloWorldPipeline = `
[[inlets.cpu]]
	interval = "3s"
	count = 1
	totalcpu = true
	percpu = false
[[outlets.file]]
	format = "json"
	decimal = 2
`
```

Run the example server.

```sh
$ go run ./httpsvr.go

start server http://127.0.0.1:8080
```

Invoke the endpoint.

```sh
$ curl -o - http://127.0.0.1:8080/helloworld

{"total_percent":8.95}
```