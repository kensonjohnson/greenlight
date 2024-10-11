package vcs

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	var revision string
	var modified bool

	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					modified = true
				}
			}
		}
	}

	if modified {
		return fmt.Sprintf("%s-dirty", revision)
	}

	return revision
}
