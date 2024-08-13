package chrome_test

import (
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/image"
	_ "github.com/OutOfBedlam/tine/x/chrome"
)

func TestChromeSnapFlow(t *testing.T) {
	dsl := `
		[[inlets.file]]
			data = [
				'{"url":"https://tine.thingsme.xyz", "dst_path":"../../tmp/chrome_snap_tine_docs.png"}', 
				'{"url":"https://github.com/OutOfBedlam/tine", "dst_path":"../../tmp/chrome_snap_tine_github.png"}', 
			]
			format = "json"
		[[flows.chrome_snap]]
			url_field = "url"
			out_field = "snap"
			timeout = "15s"
			parallelism = 2
			user_device = "iphone x"
			# full_page = 100
		[[outlets.image]]
			path_field = "dst_path"
			image_fields = ["snap"]
			overwrite = true
		`
	// Make the output timestamp deterministic, so we can compare it
	// This line is required only for testing
	count := int64(0)
	engine.Now = func() time.Time { count++; return time.Unix(1721954797+count, 0) }
	// Create a new pipeline
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	// Run the pipeline
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
}
