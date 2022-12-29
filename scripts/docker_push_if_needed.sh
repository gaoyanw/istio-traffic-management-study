#!/bin/bash

set -e

project=$1
binary=$2
tag=$3

existing_tags=$(gcloud container images list-tags --filter="tags:$tag" --format=json gcr.io/$project/$binary)

if [[ "$existing_tags" == "[]" ]]; then
  printf "build and push \"$binary\""

  docker build -f ./cmd/$binary/Dockerfile . -t gcr.io/$project/$binary:$tag
  docker push gcr.io/$project/$binary:$tag
else
  printf "tag exists\n"
fi

