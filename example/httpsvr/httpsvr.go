package main

import (
	"net/http"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/all"
)

func main() {
	var helloWorldPipeline = `
	[[inlets.cpu]]
		interval = "3s"
		count = 1
		totalcpu = true
		percpu = true
	[[outlets.file]]
		format = "json"
	`
	var screenshotPipeline = `
	[[inlets.screenshot]]
	   count = 1
	   displays = [0]
	[[outlets.image]]
		path = "nonamed.png"   ## <- Required to set image type (.jpeg, .png, .gif ....)
	`

	router := http.NewServeMux()
	router.HandleFunc("GET /helloworld", engine.HttpHandleFunc(helloWorldPipeline))
	router.HandleFunc("GET /screenshot", engine.HttpHandleFunc(screenshotPipeline))
	http.ListenAndServe("127.0.0.1:8080", router)
}
