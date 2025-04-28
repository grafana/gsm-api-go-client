// Copyright (C) 2025 Grafana Labs.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"
)

// WithBearerAuth adds the Authorization header with a Bearer token to each request.
func WithBearerAuth(token string) ClientOption {
	return WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer "+token)

		return nil
	})
}

// WithAcceptJSON adds the Accept: application/json header to each request.
func WithAcceptJSON() ClientOption {
	return WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Add("Accept", "application/json")

		return nil
	})
}
