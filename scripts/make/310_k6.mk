ifeq ($(CI),true)
XK6 ?= xk6
endif

ifeq ($(origin XK6),undefined)
XK6 ?= ./scripts/docker-run xk6
endif

define build_k6_template
BUILD_K6_TARGETS += build-k6-$(1)-$(2)

build-k6-$(1)-$(2) : GOOS := $(1)
build-k6-$(1)-$(2) : GOARCH := $(2)
build-k6-$(1)-$(2) : DIST_FILENAME := $(DISTDIR)/$(1)-$(2)/k6

endef

$(foreach BUILD_PLATFORM,$(PLATFORMS), \
		$(eval $(call build_k6_template,$(word 1,$(subst /, ,$(BUILD_PLATFORM))),$(word 2,$(subst /, ,$(BUILD_PLATFORM))),$(CMD))))

.PHONY: $(BUILD_K6_TARGETS)
$(BUILD_K6_TARGETS) : build-k6-% :
	$(S) echo 'Building k6 ($(GOOS)-$(GOARCH))'
	$(S) mkdir -p $(DISTDIR)/$(GOOS)-$(GOARCH)
	$(V) $(XK6) build --with github.com/grafana/gsm-api-go-client=. --output '$(DIST_FILENAME)' --os '$(GOOS)' --arch '$(GOARCH)' --build-flags '-trimpath'
	$(V) test '$(GOOS)' = '$(HOST_OS)' -a '$(GOARCH)' = '$(HOST_ARCH)' && \
		cp -a '$(DIST_FILENAME)' '$(DISTDIR)/$(notdir $(DIST_FILENAME))' || \
		true

.PHONY: k6
k6: $(BUILD_K6_TARGETS)
	$(S) echo Done.

build: k6

build-native: build-k6-$(HOST_OS)-$(HOST_ARCH)
