package main

import (
	"fmt"
	"net/http"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/all"
)

func main() {
	addr := "127.0.0.1:8080"
	// start data collector that save metrics to sqlite
	collect, _ := engine.New(engine.WithConfig(collectorPipeline))
	collect.Start()

	router := http.NewServeMux()
	router.HandleFunc("GET /", getView)
	router.HandleFunc("GET /query", HttpHandler(queryPipeline))
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
		<head>
			<title>SQLite Graph Web</title>
			<script src="https://cdn.jsdelivr.net/npm/echarts@5.5.1/dist/echarts.min.js"></script>
		</head>
		<body>
			<div id="main" style="width:600px;height:400px;"></div>
			<script type="text/javascript">
				var myChart = echarts.init(document.getElementById('main'));
				var option = {};
				option && myChart.setOption(option);
				
				function refresh_graph() {
					fetch('/query', {method: 'GET'})
					.then( rsp => rsp.json() )
					.then(function(data){
						console.log(data);
						myChart.setOption(data);
					}).catch(function(err){
						console.error(err);
					});
				}
				setInterval(refresh_graph, 5000);
			</script>
		</body>
	</html>`))
}

// pipeline config with Go Template using URL Query parameters
const queryPipeline = `
name = "query"
[log]
	path = "-"
	level = "debug"
[[inlets.sqlite]]
	count = 1
	path = "file::memdb?mode=memory&cache=shared"
    actions = [
        ['''SELECT time, name, value 
			FROM test
			WHERE
				datetime(time, 'unixepoch') > datetime('now', '-5 minutes')
			AND name = 'cpu_total_percent'
			ORDER BY time
		'''],
    ]
[[outlets.template]]
	path = "-"
	content_type = "application/json"
	column_mode = true
	timeformat = "15:04:05"
	lazy = true
	decimal = 2
	templates = ['''
		{
			"title": {
				"left": "center",
				"text": "SQLite Graph Web"
			},
			"xAxis": {
				"type": "category",
				"data": {{ json .time }}
			},
			"yAxis": {
				"type": "value"
			},
			"series": [
				{
					"data": {{ json .value }},
					"type": "line"
				}
			]
		}
	''']
`

const collectorPipeline = `
name = "collector"
[[inlets.load]]
	loads = ["load1", "load5", "load15"]
	interval = "5s"
[[inlets.cpu]]
	percpu = false
	totalcpu = true
	interval = "5s"
[[flows.flatten]]
[[outlets.sqlite]]
	path = "file::memdb?mode=memory&cache=shared"
	inits = [
		"CREATE TABLE test (time INTEGER, name TEXT, value REAL, UNIQUE(time, name))",
	]
	actions = [
		["INSERT INTO test (time, name, value) VALUES (?, ?, ?)", "_ts", "name", "value"],
	]
`
