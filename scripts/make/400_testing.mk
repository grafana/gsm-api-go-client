##@ Testing

GO_TEST_ARGS ?= $(GO_PKGS)

TEST_OUTPUT := $(DISTDIR)/test

ifeq ($(CI),true)
GOTESTSUM ?= gotestsum
endif

ifeq ($(origin GOTESTSUM),undefined)
GOTESTSUM ?= ./scripts/docker-run gotestsum
endif

.PHONY: test-go
test-go: export CGO_ENABLED=1 # Required so that -race works.
test-go: ## Run Go tests.
	$(S) echo "test backend"
	$(S) mkdir -p '$(DISTDIR)'
	$(GOTESTSUM) \
		--format standard-verbose \
		--jsonfile $(TEST_OUTPUT).json \
		--junitfile $(TEST_OUTPUT).xml \
		-- \
		$(GO_BUILD_MOD_FLAGS) \
		-cover \
		-coverprofile=$(TEST_OUTPUT).cov \
		-race \
		$(GO_TEST_ARGS)
	$(S) $(ROOTDIR)/scripts/report-test-coverage $(TEST_OUTPUT).cov

.PHONY: test
test: test-go ## Run all tests.
