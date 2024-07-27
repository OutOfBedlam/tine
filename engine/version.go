package engine

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var (
	tineVersion = ""
	tineSha     = ""
	timeVersion = ""
	goVersion   = ""
)

type Version struct {
	Major  int    `json:"major"`
	Minor  int    `json:"minor"`
	Patch  int    `json:"patch"`
	GitSHA string `json:"git"`
}

var _version *Version

func GetVersion() *Version {
	if _version == nil {
		v, err := semver.NewVersion(tineVersion)
		if err != nil {
			_version = &Version{}
		} else {
			_version = &Version{
				Major:  int(v.Major()),
				Minor:  int(v.Minor()),
				Patch:  int(v.Patch()),
				GitSHA: tineSha,
			}
		}
	}
	return _version
}

func DisplayVersion() string {
	return strings.TrimPrefix(tineVersion, "v")
}

func VersionString() string {
	// this version string output is used by the version command
	// and it is used by post-install test.
	//
	// test $(tine version | cut -d" " -f 1) = {{version}}
	//
	return fmt.Sprintf("%s %v %v", strings.TrimPrefix(tineVersion, "v"), tineSha, timeVersion)
}

func BuildCompiler() string {
	return goVersion
}

func BuildTimestamp() string {
	return timeVersion
}
