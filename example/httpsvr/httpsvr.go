package main

import (
	"fmt"
	"net/http"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/all"
)

func main() {
	addr := "127.0.0.1:8080"
	router := http.NewServeMux()
	router.HandleFunc("GET /cpu", engine.HttpHandleFunc(cpuPipeline))
	router.HandleFunc("GET /screenshot", engine.HttpHandleFunc(screenshotPipeline))
	router.HandleFunc("GET /template", engine.HttpHandleFunc(templatePipeline))

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

const screenshotPipeline = `
[[inlets.screenshot]]
   count = 1
   displays = [0]
[[outlets.image]]
	path = "nonamed.png"   ## <- Required to set image type (.jpeg, .png, .gif ....)
`

const templatePipeline = `
[[inlets.load]]
	count = 1
[[outlets.template]]
	templates = [
		'''
			<html>
			<body>
				{{ range . }}
					<li> <b>{{ (index ._ts).Format "2006 Jan 02 15:04:05" }}</b> {{ index .load1 }}
				{{ end }}
			</body>
			</html>
		''',
	]
`
