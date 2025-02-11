LOCAL_ACTIONLINT := actionlint

ifeq ($(CI),true)
ACTIONLINT ?= $(LOCAL_ACTIONLINT)
endif

ifeq ($(origin ACTIONLINT),undefined)
ACTIONLINT ?= $(ROOTDIR)/scripts/docker-run '$(LOCAL_ACTIONLINT)'
endif

.PHONY: lint-workflows
lint: lint-workflows
lint-workflows: ## Lint GitHub Actions workflows.
	$(S) echo 'Linting GitHub Actions workflows...'
	$(S) $(ROOTDIR)/scripts/validate-workflows
	$(S) $(ACTIONLINT)
	$(S) echo 'done.'
