// Copyright (C) 2025 Grafana Labs.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"net/http"
	"os"

	gsmClient "github.com/grafana/gsm-api-go-client"
)

func main() {
	const serverAddress = "http://localhost:3000"

	token := os.Getenv("GSM_API_TOKEN")
	if token == "" {
		log.Fatal("GSM_API_TOKEN is required")
	}

	apiClient, err := gsmClient.NewClientWithResponses(
		serverAddress,
		gsmClient.WithBearerAuth(token),
		gsmClient.WithAcceptJSON(),
	)
	if err != nil {
		log.Fatalf("Cannot create client: %s", err)
	}

	ctx := context.Background()

	secretValue := "super-secret"

	resp, err := apiClient.AddSecretWithResponse(ctx, gsmClient.AddSecretJSONRequestBody{
		Name:        "my-secret",
		Description: "This is a secret",
		Labels:      nil,
		Plaintext:   &secretValue,
	})

	switch {
	case err != nil:
		log.Fatalf("Cannot add secret: %s", err)

	case resp.HTTPResponse.StatusCode == http.StatusCreated:
		// The secret was created, so JSON201 is populated.
		log.Println("Secret ID:", resp.JSON201.Uuid)

	default:
		log.Fatalf("Cannot add secret: %s", resp.HTTPResponse.Status)
	}

	// do something with the response

	_ = resp
}
