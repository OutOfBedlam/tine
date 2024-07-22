package main

import (
	"net/http"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/all"
)

func main() {
	// start data collector that save metrics to sqlite
	collect, _ := engine.New(engine.WithConfig(collectorPipleine))
	collect.Start()

	router := http.NewServeMux()
	router.HandleFunc("GET /", getView)
	router.HandleFunc("GET /query", engine.HttpHandleFunc(queryPipeline))
	http.ListenAndServe("127.0.0.1:8080", router)

	// stop data collector
	collect.Stop()
}

func getView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<html>
		<body>
			<li><a href="/query?format=json&names=load_load1&names=load_load5&names=load_load15&pname=load-json">Load - json</a></li>
			<li><a href="/query?format=csv&names=load_load1&names=load_load5&names=load_load15&pname=load-csv">Load - csv</a></li>
			<li><a href="/query?format=json&names=load_load1&names=cpu_total_percent&pname=cpu-json">CPU - json</a></li>
			<li><a href="/query?format=csv&names=load_load1&names=cpu_total_percent&pname=cpu-csv">CPU - csv</a></li>
		</body>
		<!--
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
		-->
	</html>`))
}

// pipeline config with Go Template using URL Query parameters
const queryPipeline = `
name = "sqlite-{{ .pname }}"
[log]
	level = "debug"
[[inlets.sqlite]]
	count = 1
    path = "file::memory:?mode=memory&cache=shared"
    actions = [
        [
            """SELECT time, name, value FROM metrics
				WHERE name in ( {{ range $i, $v := .names }} {{if $i}},{{end}}'{{ $v }}'{{ end }} )
				ORDER BY time LIMIT 600
            """,
        ],
    ]
[[outlets.file]]
	format="{{ .format }}"
`

const collectorPipleine = `
name = "collector"
[defaults]
	interval = "1s"
[[inlets.load]]
[[inlets.cpu]]
	percpu = false
	totalcpu = true
[[flows.flatten]]
[[outlets.sqlite]]
    path = "file::memory:?mode=memory&cache=shared"
    inits = [
        """
            CREATE TABLE IF NOT EXISTS metrics (
                time INTEGER, name TEXT, value REAL, UNIQUE(time, name)
            )
        """,
    ]
    actions = [
        [
            "INSERT INTO metrics (time, name, value) VALUES (?, ?, ?)",
            "_ts", "name", "value"
        ],
    ]
`
