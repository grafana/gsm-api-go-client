// Copyright (C) 2025 Grafana Labs.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"path"
	"runtime/debug"
	"sync"
)

var version = "v0.0.0"

func Short() string {
	// We don't use "vcs.version" because that depends on building the binary in a particular way.
	return version
}

func Commit() string {
	return getBuildInfoByKey("vcs.revision")
}

func Buildstamp() string {
	return getBuildInfoByKey("vcs.time")
}

func Name() string {
	bi := getBuildInfo()
	if bi == nil {
		return "unknown"
	}

	return path.Base(bi.Path)
}

func getBuildInfoByKey(key string) string {
	bi := getBuildInfo()
	if bi == nil || len(bi.Settings) == 0 {
		return "unknown"
	}

	for _, setting := range bi.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}

	return ""
}

//nolint:gochecknoglobals // This is accessed thru other functions in this file.
var getBuildInfo = sync.OnceValue(func() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}

	return bi
})
