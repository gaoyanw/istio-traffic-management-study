GCP_PROJECT ?= $(shell gcloud config get-value core/project)
PROTOC_GO_PLUGIN := $(shell which protoc-gen-go)
TAG ?= $(shell git describe --always --tags --dirty)

PROTOS := $(shell find pkg -name "*.proto")
DESCRIPTORS := $(PROTOS:.proto=.pb)
PROTO_OUTS := $(PROTOS:.proto=.pb.go)

docker-%:
	./scripts/docker_push_if_needed.sh $(GCP_PROJECT) $* $(TAG)

docker: docker-httpserver docker-extprocserver docker-bookstoreserver

helm: docker helm-extprocserver helm-httpserver helm-bookstoreserver

helm-%:
	helm upgrade -i $* -n $* manifests/$* --set image.tag=$(TAG) --create-namespace

protos: $(PROTO_OUTS)

%.pb.go: %.proto
	protoc \
		--plugin=$(PROTOC_GO_PLUGIN) \
		-I third_party/github.com/googleapis/googleapis/ \
		-I pkg/ \
		--go_out=. \
		--go-grpc_out=require_unimplemented_servers=false:. \
		$<

descriptors: $(DESCRIPTORS)

%.pb: %.proto
	protoc \
		-I third_party/github.com/googleapis/googleapis/ \
		-I pkg/ \
		--include_source_info \
		--go-grpc_out=. \
		--include_imports \
		--descriptor_set_out=descriptors/$(*F).pb \
		$<

