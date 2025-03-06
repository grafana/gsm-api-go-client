# Grafana Secrets Management Go Client for k6

This repository provides a Go client for the Grafana Secrets Management API, along with a k6 extension for accessing secrets.

## Building the k6 Extension

To build k6 with this extension, use [xk6](https://github.com/grafana/xk6):

```bash
# Install xk6
go install go.k6.io/xk6/cmd/xk6@latest

# Build k6 with the Grafana Secrets Management extension
xk6 build --with github.com/grafana/gsm-api-go-client
```

## Using the Extension in k6 Tests

After building k6 with the extension, you can access Grafana Secrets in your k6 tests using the `--secret-source` flag:

```bash
# Run a k6 test with access to Grafana Secrets
k6 run --secret-source=grafanasecrets=url=<base64-encoded-url>:token=<path-to-token-file> script.js
```

### Parameters

- `url`: Base64-encoded URL of the Grafana Secrets Management API 
- `token`: Path to a file containing the API token for authentication

### Example: Encoding the URL

```bash
# Encode the API URL using base64 URL encoding
URL="https://your-grafana-secrets-api.example.com"
ENCODED_URL=$(echo -n $URL | base64 | tr '+/' '-_')

k6 run --secret-source=grafanasecrets=url=$ENCODED_URL:token=/path/to/token.txt script.js
```