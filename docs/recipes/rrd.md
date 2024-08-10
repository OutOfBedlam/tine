# RRD

{% hint style="info" %}
rrd plugins requires Tine to be built with `-tags rrd` which need `librrd-dev` package to be installed in advance.
{% endhint %}

## Example Web + RRD

Let's make a example application that collects CPU usage and system load average, saves the data into RRD, serves web page shows the collected data as graph.

This code starts a simple web server that serves at `http://127.0.0.1:8080` with three endpoints `/`, `/graph/load`, and `/graph/cpu`.

### Code

Please find the full source from the [github repository](https://github.com/OutOfBedlam/tine/tree/main/example/rrd_graph_web).

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/all"
	_ "github.com/OutOfBedlam/tine/x/rrd"
)

func main() {
	addr := "127.0.0.1:8080"
	// start data collector that save metrics to rrd file
	collect, _ := engine.New(engine.WithConfig(collectorPipeline))
	collect.Start()

	router := http.NewServeMux()
	router.HandleFunc("GET /", getView)
	router.HandleFunc("GET /graph/load", HttpHandler(graphLoadPipeline))
	router.HandleFunc("GET /graph/cpu", HttpHandler(graphCpuPipeline))
	fmt.Printf("\nlistener start at http://%s\n", addr)
	http.ListenAndServe(addr, router)

	// stop data collector
	collect.Stop()
}
```

### Run

Create this code as `rrd_graph_web.go` and run with `-tags rrd` build flags.

```sh
go run --tags rrd ./rrd_graph_web.go
```

Then open web browser, it will shows the system's load average and cpu usage in graphs.

<figure><img src="../.gitbook/assets/web-rrd.png" alt="" width="563"><figcaption><p>RRDGraph</p></figcaption></figure>


### How this works

#### Start collecting pipeline

Make a pipeline that collecting CPU usage and load average and merge them into a record then writing on the RRD file.

```go
collect, _ := engine.New(engine.WithConfig(collectorPipeline))
collect.Start()
...omit...
collect.Stop()
```

The `collectorPipeline` is defined as:

```toml
name = "rrd-collector"
[defaults]
	interval = "1s"
[[inlets.load]]
[[inlets.cpu]]
	percpu = false
	totalcpu = true
[[flows.merge]]
	wait_limit = "2s"
[[outlets.rrd]]
	path = "./tmp/rrdweb.rrd"
	step = "1s"
	heartbeat = "2s"
	fields = [
		{ field="load.load1", ds="load1", dst="GAUGE", min=0.0, max="U" },
		{ field="load.load5", ds="load5", dst="GAUGE", min=0.0, max="U" },
		{ field="load.load15", ds="load15", dst="GAUGE", min=0.0, max="U" },
		{ field="cpu.total_percent", ds="cpu_total", dst="GAUGE", min=0.0, max="U" },		 
    ]
	rra = [
        { cf = "AVERAGE", steps = "1s", rows="3h" },
        { cf = "AVERAGE", steps = "1m", rows="3d" },
        { cf = "AVERAGE", steps = "1h", rows="30d" },
        { cf = "AVERAGE", steps = "1d", rows="13M" },
	]
```


#### Rendering HTML

Attach `getView` handler on the `/` endpoint.
```go
	router.HandleFunc("GET /", getView)
```

The `getView()` handler embeds two `<img>` elements that runs on endpoints, `/graph/load` and `/graph/cpu`, and reloads them every 2 seconds.

```go
func getView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<html>
		<body>
            <img id="rrd_load" src='/graph/load' />
            <img id="rrd_cpu" src='/graph/cpu' />
		</body>
		<script type="text/javascript">
			function refresh_load_graph() {
				document.getElementById('rrd_load').src = '/graph/load?t=' + new Date().getTime();
			}
			setInterval(refresh_load_graph, 2000);

			function refresh_cpu_graph() {
				document.getElementById('rrd_cpu').src = '/graph/cpu?t=' + new Date().getTime();
			}
			setInterval(refresh_cpu_graph, 2000);
		</script>
	</html>`))
}
```

#### HttpHandler wrapper for a pipeline

Running a pipeline as a HTTP Handler requires preparation, the below `HttpHandler()` code shows how to wrap your pipeline for a handle.

```go
func HttpHandler(config string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, _ := engine.New(
			engine.WithConfig(config),
			engine.WithWriter(w),
		)
		if err := p.Run(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.Stop()
	}
}
```

#### Attach pipelines to the endpoints

```go
	router.HandleFunc("GET /graph/load", HttpHandler(graphLoadPipeline))
	router.HandleFunc("GET /graph/cpu", HttpHandler(graphCpuPipeline))
```

The pipeline definitions of `graphLoadPipeline` and `graphCpuPipeline` are:

**graphLoadPipeline**

```toml
name = "rrdweb-load"
[[inlets.rrd_graph]]
	path = "./tmp/rrdweb.rrd"
	count = 1
	title = "System Load"
	range = "10m"
	size = [600, 200]
	theme = "gchart2"
	fields = [
		{ ds = "load1", cf="AVERAGE", type="line", name="%-6s", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
		{ ds = "load5", cf="AVERAGE", type="line", name="%-6s", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
		{ ds = "load15", cf="AVERAGE", type="area", name="%-6s", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
	]
[[outlets.image]]
	path = "nonamed.png"
```

**graphCpuPipeline**

```toml
name = "rrdweb-cpu"
[[inlets.rrd_graph]]
	path = "./tmp/rrdweb.rrd"
	count = 1
	title = "CPU Usage (%)"
	range = "10m"
	size = [600, 200]
	theme = "gchart2"
	fields = [
		{ ds = "cpu_total", cf="AVERAGE", type="line", name="%-6s", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
	]
[[outlets.image]]
	path = "nonamed.png"
```
