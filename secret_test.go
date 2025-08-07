// Copyright (C) 2025 Grafana Labs.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testLimiter is an implementation of the rate.Limiter interface. It doesn't
// impose any actual rate limiting, but it allows to test the client side.
type testLimiter struct{}

func (t testLimiter) Wait(_ context.Context) error {
	return nil
}

var _ limiter = testLimiter{}

func TestGrafanaSecretsGet(t *testing.T) {
	t.Parallel()

	const (
		secretName  = "test-secret-id"
		secretValue = "test-secret-value"
		testToken   = "test-token"
	)

	setupServer := func(t *testing.T) *httptest.Server {
		t.Helper()

		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("Received headers: %+v", r.Header)

			if r.Header.Get("Authorization") != "Bearer "+testToken {
				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			secret := DecryptedSecret{
				Plaintext: secretValue,
			}

			err := json.NewEncoder(w).Encode(secret)
			if err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}))
	}

	setupClient := func(serverURL string) *Client {
		c, _ := NewClient(serverURL, WithBearerAuth(testToken))

		return c
	}

	t.Run("successful get", func(t *testing.T) {
		t.Parallel()

		server := setupServer(t)

		defer server.Close()

		grafanaSecrets := &grafanaSecrets{
			client:  setupClient(server.URL),
			limiter: testLimiter{},
		}

		actual, err := grafanaSecrets.Get(secretName)
		require.NoError(t, err)
		require.Equal(t, secretValue, actual)
	})

	t.Run("with rate limit", func(t *testing.T) {
		t.Parallel()

		server := setupServer(t)

		defer server.Close()

		grafanaSecrets := &grafanaSecrets{
			client:  setupClient(server.URL),
			limiter: newLimiter(defaultRequestsPerMinuteLimit, defaultRequestsBurst),
		}

		actual, err := grafanaSecrets.Get(secretName)
		require.NoError(t, err)
		require.Equal(t, secretValue, actual)
	})

	t.Run("hit rate limit", func(t *testing.T) {
		t.Parallel()

		server := setupServer(t)

		defer server.Close()

		grafanaSecrets := &grafanaSecrets{
			client:  setupClient(server.URL),
			limiter: newLimiter(120, 1),
		}

		// Loop continuously for a full second, so that we exhaust the
		// rate limit. This is assuming that the test will be able to
		// make more requests than the burst limit in that time.
		timer := time.After(1 * time.Second)

		count := 0

	LOOP:
		for {
			select {
			case <-timer:
				break LOOP

			default:
				actual, err := grafanaSecrets.Get(secretName)
				require.NoError(t, err)
				require.Equal(t, secretValue, actual)
				count++
			}
		}

		// After a second, we should have hit the rate limit and then
		// some (because the bucket keeps refilling), but we cannot go
		// past it by a significant margin.
		require.GreaterOrEqual(t, count, 2)
		require.LessOrEqual(t, count, 3)

		t.Log("Total requests made:", count)
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

func TestGetConfig(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		configData   string
		expectErr    bool
		expectConfig extConfig
	}{
		"valid config": {
			configData: `{
				"url":"http://localhost",
				"token":"test-token",
				"requestsPerMinuteLimit":100,
				"requestsBurst":10
			}`,
			expectErr: false,
			expectConfig: extConfig{
				URL:                    "http://localhost",
				Token:                  "test-token",
				RequestsPerMinuteLimit: valToPtr(100),
				RequestsBurst:          valToPtr(10),
			},
		},

		"missing URL": {
			configData: `{
				"token":"test-token"
			}`,
			expectErr: true,
		},

		"missing token": {
			configData: `{
				"url":"http://localhost"
			}`,
			expectErr: true,
		},

		"invalid JSON": {
			configData: `{
				"url":"http://localhost",
				"token":"test-token",
			`, // Missing closing brace
			expectErr: true,
		},

		"empty config": {
			configData: `{}`,
			expectErr:  true,
		},

		"negative requestsPerMinuteLimit": {
			configData: `{
				"url":"http://localhost",
				"token":"test-token",
				"requestsPerMinuteLimit":-100,
				"requestsBurst":10
			}`,
			expectErr: true,
		},

		"negative requestsBurst": {
			configData: `{
				"url":"http://localhost",
				"token":"test-token",
				"requestsPerMinuteLimit":100,
				"requestsBurst":-10
			}`,
			expectErr: true,
		},

		"zero requestsBurst": {
			configData: `{
				"url":"http://localhost",
				"token":"test-token",
				"requestsPerMinuteLimit":100,
				"requestsBurst":0
			}`,
			expectErr: true,
		},

		"zero requestsPerMinuteLimit": {
			configData: `{
				"url":"http://localhost",
				"token":"test-token",
				"requestsPerMinuteLimit":0,
				"requestsBurst":10
			}`,
			expectErr: true,
		},

		"missing requestsPerMinuteLimit": {
			configData: `{
				"url":"http://localhost",
				"token":"test-token",
				"requestsBurst":10
			}`,
			expectErr: false,
			expectConfig: extConfig{
				URL:                    "http://localhost",
				Token:                  "test-token",
				RequestsPerMinuteLimit: valToPtr(defaultRequestsPerMinuteLimit),
				RequestsBurst:          valToPtr(10),
			},
		},

		"missing requestsBurst": {
			configData: `{
				"url":"http://localhost",
				"token":"test-token",
				"requestsPerMinuteLimit":100
			}`,
			expectErr: false,
			expectConfig: extConfig{
				URL:                    "http://localhost",
				Token:                  "test-token",
				RequestsPerMinuteLimit: valToPtr(100),
				RequestsBurst:          valToPtr(defaultRequestsBurst),
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.json")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(testcase.configData)
			require.NoError(t, err)
			tmpFile.Close()

			configArg := "config=" + tmpFile.Name()
			config, err := getConfig(configArg)

			if testcase.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.expectConfig, config)
			}
		})
	}
}

func valToPtr[T any](v T) *T { return &v }
