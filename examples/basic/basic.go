package main

import (
	"context"
	"log"
	"net/http"
	"os"

	gsmClient "github.com/grafana/gsm-api-go-client"
)

func main() {
	token := os.Getenv("GSM_API_TOKEN")
	if token == "" {
		log.Fatal("GSM_API_TOKEN is required")
	}

	client, err := gsmClient.NewClient("http://localhost:3000", withAuth(token))
	if err != nil {
		log.Fatalf("Cannot create client: %s", err)
	}

	ctx := context.Background()

	secretValue := `super-secret`

	resp, err := client.AddSecret(ctx, gsmClient.AddSecretJSONRequestBody{
		Name:        "my-secret",
		Description: "This is a secret",
		Plaintext:   &secretValue,
		Labels:      nil,
	})
	if err != nil {
		log.Fatalf("Cannot add secret: %s", err)
	}

	defer resp.Body.Close()

	// do something with the response

	_ = resp
}

func withAuth(token string) gsmClient.ClientOption {
	addToken := func(_ context.Context, req *http.Request) error {
		req.Header.Add("Authorization", "Bearer "+token)

		return nil
	}

	return gsmClient.WithRequestEditorFn(addToken)
}
