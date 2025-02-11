LOCAL_LINT_PROTO := $(ROOTDIR)/scripts/lint-proto

ifeq ($(CI),true)
LINT_PROTO ?= $(LOCAL_LINT_PROTO)
endif

ifeq ($(origin LINT_PROTO),undefined)
LINT_PROTO ?= $(ROOTDIR)/scripts/docker-run '$(LOCAL_LINT_PROTO)'
endif

.PHONY: lint-proto
lint: lint-proto
lint-proto: ## Lint protobuf definitions.
	$(S) echo 'Linting protobuf code...'
ifeq ($(HAS_PROTO),true)
	$(S) $(LINT_PROTO)
	$(S) echo 'done.'
else
	$(S) echo 'No protobuf files found.'
endif
