LOCAL_GEN_PROTO ?= $(ROOTDIR)/scripts/gen-proto

ifeq ($(CI),true)
GEN_PROTO ?= '$(LOCAL_GEN_PROTO)'
endif

ifeq ($(origin GEN_PROTO),undefined)
GEN_PROTO ?= ./scripts/docker-run '$(LOCAL_GEN_PROTO)'
endif

.PHONY: generate-protobuf
generate: generate-protobuf
generate-protobuf: ## Generate protobuf code.
ifeq ($(HAS_PROTO),true)
	$(S) echo 'Generating protobuf code...'
	$(V) $(GEN_PROTO)
	$(S) echo 'Done.'
else
	$(S) echo 'No protobuf files found.'
endif
