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

	c, err := client.NewClient("http://localhost:3000", withAuth(token))
	if err != nil {
		log.Fatalf("Cannot create client: %s", err)
	}

	ctx := context.Background()

	resp, err := c.AddSecret(ctx, client.AddSecretJSONRequestBody{
		Name:        "my-secret",
		Description: "This is a secret",
		Plaintext:   "super-secret",
		Labels:      nil,
	})
}

func withAuth(token string) client.ClientOption {
	addToken := func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer "+token)
		return nil
	}

	return client.WithRequestEditorFn(addToken)
}
