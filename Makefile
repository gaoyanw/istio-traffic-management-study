GCP_PROJECT ?= $(shell gcloud config get-value core/project)
TAG ?= $(shell git describe --always --tags --dirty)

docker-%:
	./scripts/docker_push_if_needed.sh $(GCP_PROJECT) $* $(TAG)

docker: docker-httpserver docker-extprocserver

helm-extprocserver:
	helm upgrade -i extprocserver -n extprocserver manifests/extprocserver \
		--set image.tag=$(TAG)

