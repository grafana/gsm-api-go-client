// Copyright (C) 2025 Grafana Labs.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAll validates that the various function are returning the expected
// values. They exist mostly to make sure that they aren't modified without
// some thought.
func TestAll(t *testing.T) {
	t.Parallel()
	require.Equal(t, version, Short())
	require.NotEmpty(t, Commit())
	require.NotEmpty(t, Buildstamp())
}
