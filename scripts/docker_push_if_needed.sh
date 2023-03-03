#!/bin/bash

set -e

project=$1
binary=$2
tag=$3
echo "the current image tag is: $tag"
  docker build -f ./cmd/$binary/Dockerfile . -t gcr.io/$project/$binary:$tag
  docker push gcr.io/$project/$binary:$tag

