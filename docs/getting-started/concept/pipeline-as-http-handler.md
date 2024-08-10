# Pipeline as Http Handler

A pipeline can serves as a http handler with Go standard `net/http`.

The following example code demonstrates a simple web service that retrieves and returns the CPU usage of the system.

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
	router.HandleFunc("GET /cpu", engine.HttpHandleFunc(cpuPipeline))

	fmt.Printf("\nstart server http://%s\n\n", addr)
	http.ListenAndServe(addr, router)
}

const cpuPipeline = `
[[inlets.cpu]]
	interval = "3s"
	totalcpu = true
	percpu = false
[[flows.select]]
	includes = ["#_ts", "*"]
[[outlets.file]]
	format = "json"
	decimal = 2
`
```

To run the example server, execute the following command in your terminal:

```sh
$ go run ./httpsvr.go

start server at http://127.0.0.1:8080
```

Once the server is up and running, you can make a request to the `/cpu` endpoint by treating it as a RESTful API call. Use the appropriate HTTP method and include the endpoint URL in your request.

For example, you can use the `curl` command to make a GET request to the `/cpu` endpoint:

```sh
$  curl -o - -v http://127.0.0.1:8080/cpu ↵

> GET /cpu HTTP/1.1
> Host: 127.0.0.1:8080
>
< HTTP/1.1 200 OK
< Content-Type: application/x-ndjson
< Transfer-Encoding: chunked
<
{"_ts":1723288311,"total_percent":10.59}
{"_ts":1723288314,"total_percent":2.94}
{"_ts":1723288317,"total_percent":3.11}
^C
```
