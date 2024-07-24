//go:build !windows
// +build !windows

//go:generate sh -c "go run ../tools/genconf/* > ../tine.toml"

package cmd
