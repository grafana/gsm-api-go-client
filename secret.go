// Package client implements a k6 extension for accessing Grafana Secrets Management.
// To use this extension, build k6 with xk6-build:
//
//	xk6 build --with github.com/grafana/gsm-api-go-client
package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go.k6.io/k6/secretsource"
)

func withAuth(token string) ClientOption {
	addToken := func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer "+token)
		return nil
	}

	return WithRequestEditorFn(addToken)
}

func init() {
	secretsource.RegisterExtension("grafanasecrets", func(params secretsource.Params) (secretsource.Source, error) {
		list := strings.Split(params.ConfigArgument, ":")
		r := make(map[string]string, len(list))
		for _, kv := range list {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return nil, fmt.Errorf("parsing %q, needs =", kv)
			}

			r[k] = v
		}

		encodedURL, ok := r["url"]
		if !ok {
			return nil, errors.New("url parameter is required")
		}

		// Decode the base64-encoded URL
		decodedURLBytes, err := base64.URLEncoding.DecodeString(encodedURL)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 URL: %w", err)
		}
		url := string(decodedURLBytes)

		tokenPath, ok := r["token"]
		if !ok {
			return nil, errors.New("token parameter is required")
		}

		tokenBytes, err := os.ReadFile(tokenPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read token file: %w", err)
		}
		token := strings.TrimSpace(string(tokenBytes))

		client, err := NewClient(url, withAuth(token))
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}

		return &grafanaSecrets{
			url:    url,
			token:  token,
			client: client,
		}, nil
	})
}

type grafanaSecrets struct {
	url    string
	token  string
	client *Client
}

func (gs *grafanaSecrets) Name() string {
	return "Grafana Secrets"
}

func (gs *grafanaSecrets) Description() string {
	return "Grafana secrets for k6"
}

func (gs *grafanaSecrets) Get(key string) (string, error) {
	ctx := context.Background()
	response, err := gs.client.DecryptSecretById(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get secret: status code %d", response.StatusCode)
	}

	var decryptedSecret DecryptedSecret
	if err := json.NewDecoder(response.Body).Decode(&decryptedSecret); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	defer response.Body.Close()

	return decryptedSecret.Plaintext, nil
}
