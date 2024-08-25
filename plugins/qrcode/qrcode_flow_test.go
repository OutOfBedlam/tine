package qrcode_test

import (
	"testing"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/image"
	_ "github.com/OutOfBedlam/tine/plugins/qrcode"
)

func TestQRCodeFlow(t *testing.T) {
	dsl := `
		[log]
			path = "-"
			level = "debug"
			no_color = true
		[[inlets.file]]
			data = [
				"a,https://tine.thingsme.xyz",
			]
			fields = ["area", "url"]
			types = ["string", "string"]
		[[flows.qrcode]]
			input_field = "url"
			output_field = "qrcode"
			# QRCode width should be < 256
			width = 11
			# background_transparent = true
			background_color = "#ffffff"
			foreground_color = "#000000"
			# logo image should only has 1/5 width of QRCode at most (.png or .jpeg)
			logo = "./testdata/tine_x64.png"
			halftone = "./testdata/test.jpeg"
		[[outlets.image]]
			path = "./testdata/output_a.png"
			image_fields = ["qrcode"]
			overwrite = true
	`
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		t.Fatal(err)
	}
	if err := pipeline.Run(); err != nil {
		t.Fatal(err)
	}
	pipeline.Stop()
}
