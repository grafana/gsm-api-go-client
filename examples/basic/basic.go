package main

import (
	"context"
	"log"
	"net/http"
	"os"

	client "github.com/grafana/gsm-api-go-client"
)

func main() {
	token := os.Getenv("GSM_API_TOKEN")
	if token == "" {
		log.Fatal("GSM_API_TOKEN is required")
	}

	apiClient, err := client.NewClientWithResponses("http://localhost:3000", withAuth(token), withAcceptJSON())
	if err != nil {
		log.Fatalf("Cannot create client: %s", err)
	}

	ctx := context.Background()

	secret := "super-secret"

	resp, err := apiClient.AddSecretWithResponse(ctx, client.AddSecretJSONRequestBody{
		Name:        "my-secret",
		Description: "This is a secret",
		Labels:      nil,
		Plaintext:   &secret,
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
}

func withAuth(token string) client.ClientOption {
	return client.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer "+token)

		return nil
	})
}

func withAcceptJSON() client.ClientOption {
	return client.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Add("Accept", "application/json")

		return nil
	})
}
