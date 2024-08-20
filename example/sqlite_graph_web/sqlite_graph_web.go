package main

import (
	_ "embed"
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

//go:embed sqlite_graph_web.html
var viewContent []byte

func getView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(viewContent)
}

// pipeline config with Go Template using URL Query parameters
const queryPipeline = `
name = "query"
[[inlets.sqlite]]
	count = 1
	path = "file::memdb?mode=memory&cache=shared"
    actions = [
        ['''SELECT time, cpu, load1, load5, load15 
			FROM test
			WHERE
				datetime(time, 'unixepoch') > datetime('now', '-5 minutes')
			ORDER BY time
		'''],
    ]
[[outlets.template]]
	content_type = "application/json"
	column_series = "json"
	lazy = true
	decimal = 2
	templates = ['{ "time":{{ .time }}, "cpu":{{ .cpu }}, "load1":{{ .load1 }}, "load5":{{ .load5 }}, "load15":{{ .load15 }} }']
`

const collectorPipeline = `
name = "collector"
[[inlets.load]]
	loads = [1, 5, 15]
	interval = "5s"
[[inlets.cpu]]
	percpu = false
	totalcpu = true
	interval = "5s"
[[flows.merge]]
	wait_limit = "2.5s"
[[outlets.sqlite]]
	path = "file::memdb?mode=memory&cache=shared"
	inits = [
		"""CREATE TABLE test (
			time INTEGER,
			cpu REAL,
			load1 REAL,
			load5 REAL,
			load15 REAL,
			UNIQUE(time)
		)""",
	]
	actions = [
		["""INSERT INTO test (time, cpu, load1, load5, load15) 
			VALUES (?, ?, ?, ?, ?)""", 
			"_ts", "cpu_total_percent", "load_load1", "load_load5", "load_load15"],
		["""DELETE FROM test WHERE datetime(time, 'unixepoch') < datetime('now', '-60 minutes')"""],
	]
`
