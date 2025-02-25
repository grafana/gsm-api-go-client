package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGrafanaSecretsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received headers: %+v", r.Header)
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		secret := DecryptedSecret{
			Plaintext: "test-secret-value",
		}
		json.NewEncoder(w).Encode(secret)
	}))
	defer server.Close()

	gs := &grafanaSecrets{
		url:   server.URL,
		token: "test-token",
		client: func() *Client {
			c, _ := NewClient(server.URL, withAuth("test-token"))
			return c
		}(),
	}

	t.Run("successful get", func(t *testing.T) {
		secret, err := gs.Get("test-secret-id")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if secret != "test-secret-value" {
			t.Errorf("got secret = %v, want %v", secret, "test-secret-value")
		}
	})
}
