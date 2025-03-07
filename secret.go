// Package client implements a k6 extension for accessing Grafana Secrets Management.
// To use this extension, build k6 with xk6-build:
//
//	xk6 build --with github.com/grafana/gsm-api-go-client
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go.k6.io/k6/secretsource"
)

// Config holds the configuration for Grafana Secrets
type Config struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func withAuth(token string) ClientOption {
	addToken := func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer "+token)
		return nil
	}

	return WithRequestEditorFn(addToken)
}

func ParseConfigArgument(configArg string) (string, error) {
	configKey, configPath, ok := strings.Cut(configArg, "=")
	if !ok || configKey != "config" {
		return "", fmt.Errorf("config parameter is required in format 'config=path/to/config'")
	}
	return configPath, nil
}

func init() {
	secretsource.RegisterExtension("grafanasecrets", func(params secretsource.Params) (secretsource.Source, error) {
		// Parse the ConfigArgument to get the config file path
		configPath, err := ParseConfigArgument(params.ConfigArgument)
		if err != nil {
			return nil, err
		}

		configData, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		var config Config
		if err := json.Unmarshal(configData, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}

		if config.URL == "" {
			return nil, errors.New("url is required in config file")
		}
		if config.Token == "" {
			return nil, errors.New("token is required in config file")
		}

		client, err := NewClient(config.URL, withAuth(config.Token))
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}

		return &grafanaSecrets{
			client: client,
		}, nil
	})
}

type grafanaSecrets struct {
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
