package main

import (
	"net/http"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/all"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /helloworld", getHelloWorld)
	router.HandleFunc("GET /screenshot", getScreenshot)
	http.ListenAndServe("127.0.0.1:8080", router)
}

func getHelloWorld(w http.ResponseWriter, r *http.Request) {
	var helloWorldPipeline = `
	[[inlets.cpu]]
		interval = "3s"
		count = 1
		totalcpu = true
		percpu = true
	[[outlets.file]]
		path = "-"
		format = "json"
	`
	// Create engine
	pipeline, err := engine.New(
		engine.WithName("helloworld"),
		engine.WithConfig(helloWorldPipeline),
		engine.WithWriter(w), // <-- Redirect stdout to http.ResponseWriter
		engine.WithSetContentEncodingFunc(func(contentEncoding string) {
			w.Header().Set("Content-Encoding", contentEncoding)
		}),
		engine.WithSetContentTypeFunc(func(contentType string) {
			w.Header().Set("Content-Type", contentType)
		}),
	)
	if err != nil {
		panic(err)
	}

	// Execute engine
	err = pipeline.Run()
	if err != nil {
		panic(err)
	}

	// Stop engine
	pipeline.Stop()
}

func getScreenshot(w http.ResponseWriter, r *http.Request) {
	var screenshotPipeline = `
	[[inlets.screenshot]]
	   count = 1
	   displays = [0]
	[[outlets.image]]
		path = "nonamed.png"
	`

	// Create engine
	pipeline, err := engine.New(
		engine.WithName("screenshot"),
		engine.WithConfig(screenshotPipeline),
		engine.WithWriter(w), // <-- Redirect stdout to http.ResponseWriter
		engine.WithSetContentTypeFunc(func(contentType string) {
			w.Header().Set("Content-Type", contentType)
		}),
	)
	if err != nil {
		panic(err)
	}

	// Execute engine
	err = pipeline.Run()
	if err != nil {
		panic(err)
	}

	// Stop engine
	pipeline.Stop()
}
