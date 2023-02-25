GCP_PROJECT ?= $(shell gcloud config get-value core/project)
PROTOC_GO_PLUGIN := $(shell which protoc-gen-go)
TAG ?= $(shell git describe --always --tags --dirty)

PROTOS := $(shell find pkg -name "*.proto")
DESCRIPTORS := $(PROTOS:.proto=.pb)
PROTO_OUTS := $(PROTOS:.proto=.pb.go)

docker-%:
	./scripts/docker_push_if_needed.sh $(GCP_PROJECT) $* $(TAG)

docker: docker-httpserver docker-extprocserver docker-bookstoreserver docker-bookstorehttpserver

helm: docker helm-extprocserver helm-httpserver helm-bookstoreserver helm-bookstorehttpserver

helm-%:
	helm upgrade -i $* -n $* manifests/$* \
		--set image.tag=$(TAG) \
		--set image.repository=gcr.io/$(GCP_PROJECT)/$* \
		--create-namespace

protos: $(PROTO_OUTS)

%.pb.go: %.proto
	protoc \
		--plugin=$(PROTOC_GO_PLUGIN) \
		-I third_party/github.com/googleapis/googleapis/ \
		-I pkg/ \
		--go_out=. \
		--go-grpc_out=require_unimplemented_servers=false:. \
		$<

manifests/bookstoreserver/bookstore.pb: pkg/apis/bookstore/bookstore.proto
	protoc \
		-I third_party/github.com/googleapis/googleapis/ \
		-I pkg/ \
		--include_source_info \
		--include_imports \
		--descriptor_set_out=manifests/bookstoreserver/bookstore.pb \
	 	pkg/apis/bookstore/bookstore.proto

