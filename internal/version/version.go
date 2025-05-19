// Copyright (C) 2025 Grafana Labs.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"runtime/debug"
	"slices"
	"strings"
	"sync"
)

const (
	revisionKey = "vcs.revision"
	timeKey     = "vcs.time"
)

func Short() string {
	bi := getBuildInfo()

	return bi.Main.Version
}

func Commit() string {
	return getBuildInfoByKey(revisionKey)
}

func Buildstamp() string {
	return getBuildInfoByKey(timeKey)
}

func getBuildInfoByKey(key string) string {
	buildinfo := getBuildInfo()
	if buildinfo == nil {
		return "invalid"
	}

	idx, found := slices.BinarySearchFunc(buildinfo.Settings, key, isKey)
	if found {
		return buildinfo.Settings[idx].Value
	}

	return "unknown"
}

//nolint:gochecknoglobals // This variable is only accessed in this package.
var getBuildInfo = sync.OnceValue(func() *debug.BuildInfo {
	buildinfo, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}

	slices.SortFunc(buildinfo.Settings, cmpBuildSettings)

	return buildinfo
})

func cmpBuildSettings(a, b debug.BuildSetting) int {
	if v := strings.Compare(a.Key, b.Key); v != 0 {
		return v
	}

	return strings.Compare(a.Value, b.Value)
}

func isKey(a debug.BuildSetting, key string) int {
	return strings.Compare(a.Key, key)
}
