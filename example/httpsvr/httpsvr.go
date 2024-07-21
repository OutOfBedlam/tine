package main

import (
	"net/http"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/all"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /helloworld", engine.HttpHandleFunc(helloWorldPipeline))
	router.HandleFunc("GET /screenshot", engine.HttpHandleFunc(screenshotPipeline))
	router.HandleFunc("GET /template", engine.HttpHandleFunc(templatePipeline))
	http.ListenAndServe("127.0.0.1:8080", router)
}

const helloWorldPipeline = `
[[inlets.cpu]]
	interval = "3s"
	count = 1
	totalcpu = true
	percpu = true
[[outlets.file]]
	format = "json"
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
