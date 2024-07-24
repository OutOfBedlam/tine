//go:build windows
// +build windows

//go:generate cmd.exe /c "go run tools/genconf/genconf.go > tine.toml"

package cmd

import _ "time/tzdata"
