// Code generated by go generate; DO NOT EDIT.
//go:generate go run all_gen.go

package all

import (
	// formats
	_ "github.com/OutOfBedlam/tine/plugin/codec/csv"
	_ "github.com/OutOfBedlam/tine/plugin/codec/json"

	// compressors
	_ "github.com/OutOfBedlam/tine/plugin/codec/compress"
	_ "github.com/OutOfBedlam/tine/plugin/codec/snappy"

	// inlets
	_ "github.com/OutOfBedlam/tine/plugin/inlets/exec"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/http"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/nats"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/psutil"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/screenshot"
	_ "github.com/OutOfBedlam/tine/plugin/inlets/syslog"

	// outlets
	_ "github.com/OutOfBedlam/tine/plugin/outlets/excel"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/http"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/image"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/mqtt"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/template"

	// flows
	_ "github.com/OutOfBedlam/tine/plugin/flows/base"
)

const A = ""
