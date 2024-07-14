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
	_ "github.com/OutOfBedlam/tine/plugin/inlets/syslog"

	// outlets
	_ "github.com/OutOfBedlam/tine/plugin/outlets/file"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/http"
	_ "github.com/OutOfBedlam/tine/plugin/outlets/mqtt"

	// flows
	_ "github.com/OutOfBedlam/tine/plugin/flows/cel"
	_ "github.com/OutOfBedlam/tine/plugin/flows/damper"
	_ "github.com/OutOfBedlam/tine/plugin/flows/expr"
	_ "github.com/OutOfBedlam/tine/plugin/flows/monad"
	_ "github.com/OutOfBedlam/tine/plugin/flows/name"
)

const A = ""
