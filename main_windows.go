//go:build windows
// +build windows

//go:generate cmd /c "go run tools/genconf/* > tine.toml"

package main

import _ "time/tzdata"
