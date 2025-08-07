// Copyright (C) 2025 Grafana Labs.
// SPDX-License-Identifier: Apache-2.0

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
	"time"

	"go.k6.io/k6/secretsource"
	"golang.org/x/time/rate"
)

var (
	errInvalidConfig                 = errors.New("config parameter is required in format 'config=path/to/config'")
	errMissingURL                    = errors.New("url is required in config file")
	errMissingToken                  = errors.New("token is required in config file")
	errFailedToGetSecret             = errors.New("failed to get secret")
	errInvalidRequestsPerMinuteLimit = errors.New("requestsPerMinuteLimit must be greater than 0")
	errInvalidRequestsBurst          = errors.New("requestsBurst must be greater than 0")
)

// extConfig holds the configuration for Grafana Secrets.
type extConfig struct {
	URL                    string `json:"url"`
	Token                  string `json:"token"`
	RequestsPerMinuteLimit *int   `json:"requestsPerMinuteLimit"`
	RequestsBurst          *int   `json:"requestsBurst"`
}

func ParseConfigArgument(configArg string) (string, error) {
	configKey, configPath, ok := strings.Cut(configArg, "=")
	if !ok || configKey != "config" {
		return "", errInvalidConfig
	}

	return configPath, nil
}

const (
	// The rate limiter replenishes tokens in the bucket once every 200 ms
	// (5 per second) and allows a burst of 10 requests firing faster than
	// that. If the client keeps making requests at the rapid pace, they
	// will be slowed down. This allows a client to ask for a bunch of
	// secrets at the start of a script, and then it slows it down to a
	// reasonable pace. A single script using more than 25 secrets is
	// probably a bad idea anyway. The reate limit is here to protect the
	// Grafana Secrets API from being hammered too hard by bugs in a
	// script (for example, someone using a secret inside a loop).
	//
	// These values can be adjusted as needed.
	defaultRequestsPerMinuteLimit = 300 // 300 requests per minute is one request every 200 ms
	defaultRequestsBurst          = 10  // Allow a burst of 10 requests
)

//nolint:gochecknoinits // This is how xk6 works.
func init() {
	secretsource.RegisterExtension("grafanasecrets", func(params secretsource.Params) (secretsource.Source, error) {
		config, err := getConfig(params.ConfigArgument)
		if err != nil {
			return nil, fmt.Errorf("missing or invalid config: %w", err)
		}

		client, err := NewClient(config.URL, WithBearerAuth(config.Token))
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}

		return &grafanaSecrets{
			client:  client,
			limiter: newLimiter(*config.RequestsPerMinuteLimit, *config.RequestsBurst),
		}, nil
	})
}

type grafanaSecrets struct {
	client  *Client
	limiter limiter
}

func (gs *grafanaSecrets) Name() string {
	return "Grafana Secrets"
}

func (gs *grafanaSecrets) Description() string {
	return "Grafana secrets for k6"
}

func (gs *grafanaSecrets) Get(key string) (string, error) {
	ctx := context.Background()

	if err := gs.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter error: %w", err)
	}

	response, err := gs.client.DecryptSecretById(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code %d: %w", response.StatusCode, errFailedToGetSecret)
	}

	var decryptedSecret DecryptedSecret
	if err := json.NewDecoder(response.Body).Decode(&decryptedSecret); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return decryptedSecret.Plaintext, nil
}

type limiter interface {
	Wait(ctx context.Context) error
}

func newLimiter(requestsPerMinuteLimit, requestsBurst int) *rate.Limiter {
	// The calculation below looks wrong because it seems like it's giving
	// n min²/req, but the first number is actually time unit/min, so the
	// units of the result are time unit/req, which is correct because it's
	// the interval of time after which a new token is replenished. In
	// other words, the units are time unit/token.
	tokenReplenishInterval := time.Minute / time.Duration(requestsPerMinuteLimit)

	return rate.NewLimiter(rate.Every(tokenReplenishInterval), requestsBurst)
}

func getConfig(arg string) (extConfig, error) {
	var config extConfig

	// Parse the ConfigArgument to get the config file path
	configPath, err := ParseConfigArgument(arg)
	if err != nil {
		return config, err
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(configData, &config); err != nil {
		return config, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	if config.URL == "" {
		return config, errMissingURL
	}

	if config.Token == "" {
		return config, errMissingToken
	}

	if config.RequestsPerMinuteLimit == nil {
		requestsPerMinuteLimit := defaultRequestsPerMinuteLimit
		config.RequestsPerMinuteLimit = &requestsPerMinuteLimit
	}

	if config.RequestsBurst == nil {
		requestsBurst := defaultRequestsBurst
		config.RequestsBurst = &requestsBurst
	}

	if *config.RequestsPerMinuteLimit <= 0 {
		return config, errInvalidRequestsPerMinuteLimit
	}

	if *config.RequestsBurst <= 0 {
		return config, errInvalidRequestsBurst
	}

	return config, nil
}
