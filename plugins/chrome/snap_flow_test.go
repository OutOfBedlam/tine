package chrome_test

import (
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	"github.com/OutOfBedlam/tine/plugins/chrome"
	_ "github.com/OutOfBedlam/tine/plugins/image"
)

func TestChromeSnapFlow(t *testing.T) {
	if chrome.FindExecPath() == "" {
		t.Skip("chrome_snap test is only running when google-chrome is installed")
	}
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
			device = "iPhoneX"
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
