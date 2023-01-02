#!/bin/bash

INGRESS_IP=$(kubectl get svc -n bookstoreserver bookstoreserver -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
PORT=${PORT:-8080}

SERVER=$INGRESS_IP:$PORT

echo "Server = $SERVER"

if [[ "$1" == "grpc" ]]; then
  go run cmd/bookstoreclient/main.go -server $SERVER
  exit
fi

echo "Create shelf"
create_output=$(curl -s $SERVER/v1/shelves -XPOST -H "Content-type: application/json" \
  -d '{ "theme": "Music" }')
echo $create_output

shelf_id=$(echo $create_output | jq -r .id)

echo "List shelves"
curl -s $SERVER/v1/shelves

echo "Delete shelf \"$shelf_id\""
curl -s -XDELETE $SERVER/v1/shelves/$shelf_id

echo "List shelves"
curl -s $SERVER/v1/shelves

