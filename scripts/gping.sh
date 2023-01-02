#!/bin/bash

INGRESS_IP=$(kubectl get svc -n bookstoreserver bookstoreserver -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo "INGRESS_IP=${INGRESS_IP}"
go run cmd/bookstoreclient/main.go -server ${INGRESS_IP}:8080
