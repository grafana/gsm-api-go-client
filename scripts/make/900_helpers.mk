##@ Helpers

# rwildcard will recursively search for files matching the pattern, e.g. $(call rwildcard, src/*.c)
rwildcard = $(call rwildcard_helper, $(dir $1), $(notdir $1))
rwildcard_helper = $(wildcard $(addsuffix $(strip $2), $(strip $1))) \
	    $(foreach d, $(wildcard $(addsuffix *, $(strip $1))), $(call rwildcard_helper, $d/, $2))

.PHONY: help
help: ## Display this help.
	$(S) awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean up intermediate build artifacts.
	$(S) echo "Cleaning intermediate build artifacts..."
	$(V) rm -rf node_modules
	$(V) rm -rf public/build
	$(V) rm -rf "$(DISTDIR)/build"
	$(V) rm -rf "$(DISTDIR)/publish"

.PHONY: distclean
distclean: clean ## Clean up all build artifacts.
	$(S) echo "Cleaning all build artifacts..."
	$(V) git clean -Xf

.PHONY: version
version: ## Create version information file.
	$(S) mkdir -p $(DISTDIR)
	$(S) ./scripts/version | tee $(DISTDIR)/version

.PHONY: testdata
testdata: ## Update golden files for tests.
	$(S) true

GBT_GO_VERSION_FILE := $(strip $(DISTDIR)/$(subst :,_,$(subst /,_,$(GBT_IMAGE))).go_version)

$(GBT_GO_VERSION_FILE) :
	$(ROOTDIR)/scripts/docker-run go version | cut -d ' ' -f3 | sed -e 's,^go,,' > "$@"

.PHONY: update-go-version
update-go-version: $(GBT_GO_VERSION_FILE)
update-go-version: ## Update Go version (specify in go.mod)
	$(S) $(GO) mod edit -go=$(shell cat "$(GBT_GO_VERSION_FILE)")
