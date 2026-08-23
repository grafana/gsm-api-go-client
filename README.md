# Grafana Secrets Management Go Client for k6

This repository provides a Go client for the Grafana Secrets Management API.

## API documentation

Browse the [generated API reference](https://zrfsro.github.io/gsm-api-go-client-sourcey-docs/), built from the checked-in `openapi.yaml` contract with Sourcey 3.6.5. Rebuild it locally with:

```sh
npx --yes sourcey@3.6.5 validate openapi.yaml
npx --yes sourcey@3.6.5 build openapi.yaml -o sourcey-dist
```