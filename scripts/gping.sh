#!/bin/bash

# curl the bookstore gRPC backend using two methods, first HTTP with transcoding, second use gRPC directly
INGRESS_IP=$(kubectl get svc -n bookstoreserver bookstoreserver -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
PORT=${PORT:-8080} # ${parameter:-word} if parameter is null or unset, it expands to the value of word , otherwise to the value of parameter .

SERVER=${SERVER:-$INGRESS_IP:$PORT}
echo "Server = $SERVER"

if [[ "$1" == "grpc" ]]; then
  go run cmd/bookstoreclient/main.go -server $SERVER
  exit
fi

AUTH="allow"

echo "Create shelf"
create_output=$(curl -s $SERVER/v1/shelves -XPOST -H "Content-type: application/json" \
  -H "x-ext-authz: $AUTH"\
  -d '{ "theme": "Music" }')
echo $create_output

shelf_id=$(echo $create_output | jq -r .id)

echo "List shelves"
curl -s $SERVER/v1/shelves   -H "x-ext-authz: $AUTH"

echo "Delete shelf \"$shelf_id\""
curl -s -XDELETE $SERVER/v1/shelves/$shelf_id -H "x-ext-authz: $AUTH"

echo "List shelves"
curl -s $SERVER/v1/shelves -H "x-ext-authz: $AUTH"