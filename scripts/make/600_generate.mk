##@ Code generation

ifeq ($(CI),true)
GO_GENERATE ?= go generate
endif

ifeq ($(origin GO_GENERATE),undefined)
GO_GENERATE ?= ./scripts/docker-run go generate
endif

.PHONY: generate
generate: ## Generate code.
	$(V) $(GO_GENERATE) ./...
