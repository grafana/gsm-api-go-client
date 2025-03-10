// Copyright (C) 2024 Grafana Labs.
// SPDX-License-Identifier: AGPL-3.0-only

//go:build ruleguard
// +build ruleguard

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func noContextTODO(m dsl.Matcher) {
	m.Match(`context.TODO()`).Report(`should use another context if possible, not context.TODO()`)
}
