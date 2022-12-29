GCP_PROJECT ?= $(shell gcloud config get-value core/project)
TAG ?= $(shell git describe --always --tags --dirty)

docker-%:
	./scripts/docker_push_if_needed.sh $(GCP_PROJECT) $* $(TAG)

docker: docker-httpserver docker-extprocserver

helm: docker helm-extprocserver helm-httpserver

helm-%:
	helm upgrade -i $* -n $* manifests/$* --set image.tag=$(TAG)

