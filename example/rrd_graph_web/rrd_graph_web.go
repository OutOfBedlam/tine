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

const graphLoadPipeline = `
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
`

const graphCpuPipeline = `
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
`
const collectorPipeline = `
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
`
