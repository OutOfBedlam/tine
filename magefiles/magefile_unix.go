//go:build !windows
// +build !windows

package main

import (
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

func IsMacOS13() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return false
	}
	release := string(utsname.Release[:])
	// macOS 13 : see https://en.wikipedia.org/wiki/MacOS_version_history
	return strings.HasPrefix(release, "22.")
}
