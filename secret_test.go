// Copyright (C) 2025 Grafana Labs.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGrafanaSecretsGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received headers: %+v", r.Header)

		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		secret := DecryptedSecret{
			Plaintext: "test-secret-value",
		}

		err := json.NewEncoder(w).Encode(secret)
		if err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	grafanaSecrets := &grafanaSecrets{
		client: func() *Client {
			c, _ := NewClient(server.URL, WithBearerAuth("test-token"))

			return c
		}(),
	}

	t.Run("successful get", func(t *testing.T) {
		secret, err := grafanaSecrets.Get("test-secret-id")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if secret != "test-secret-value" {
			t.Errorf("got secret = %v, want %v", secret, "test-secret-value")
		}
	})
}

func TestParseConfigArgument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configArg string
		wantPath  string
		wantErr   bool
	}{
		{
			name:      "valid config argument",
			configArg: "config=/path/to/config.json",
			wantPath:  "/path/to/config.json",
			wantErr:   false,
		},
		{
			name:      "empty config argument",
			configArg: "",
			wantErr:   true,
		},
		{
			name:      "no equals sign",
			configArg: "config",
			wantErr:   true,
		},
		{
			name:      "wrong key",
			configArg: "wrongkey=/path/to/config.json",
			wantErr:   true,
		},
	}

	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			gotPath, err := ParseConfigArgument(testcase.configArg)
			if testcase.wantErr {
				if err == nil {
					t.Errorf("ParseConfigArgument() error = nil, wantErr = true")

					return
				}

				return
			}

			if err != nil {
				t.Errorf("ParseConfigArgument() unexpected error = %v", err)

				return
			}

			if gotPath != testcase.wantPath {
				t.Errorf("ParseConfigArgument() = %q, want %q", gotPath, testcase.wantPath)
			}
		})
	}
}
