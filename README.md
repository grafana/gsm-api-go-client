# Grafana Secrets Management Go Client for k6

This repository provides a Go client for the Grafana Secrets Management API, along with a k6 extension for accessing secrets.

## Building the k6 Extension

To build k6 with this extension, use [xk6](https://github.com/grafana/xk6):

```bash
# Install xk6
go install go.k6.io/xk6/cmd/xk6@latest

# Build k6 with the latest extension
xk6 build --with github.com/grafana/gsm-api-go-client

# or build a local copy of the extension
xk6 build --with github.com/grafana/gsm-api-go-client=.
```

## Usage

After building k6 with the extension, you can access Grafana Secrets in your k6 tests using the `--secret-source` flag:

```bash
k6 run --secret-source=grafanasecrets=config=path/to/config.json script.js
```

### Config File

Create a JSON config file with your API URL and token:

```bash
echo '{"url": "https://your-grafana-secrets-api.example.com", "token": "api-token"}' | jq . > config.json

k6 run --secret-source=grafanasecrets=config=./config.json script.js
```