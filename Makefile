GCP_PROJECT ?= $(shell gcloud config get-value core/project)
TAG ?= $(shell git describe --always --tags --dirty)

docker-httpserver:
	docker build -f ./cmd/httpserver/Dockerfile . -t gcr.io/$(GCP_PROJECT)/httpserver:$(TAG)
	docker push gcr.io/$(GCP_PROJECT)/httpserver:$(TAG)

docker-extprocserver:
	docker build -f ./cmd/extprocserver/Dockerfile . -t gcr.io/$(GCP_PROJECT)/extprocserver:$(TAG)
	docker push gcr.io/$(GCP_PROJECT)/extprocserver:$(TAG)

docker: docker-httpserver docker-extprocserver

helm-extprocserver:
	helm upgrade -i extprocserver -n extprocserver manifests/extprocserver \
		--set image.tag=$(TAG)

