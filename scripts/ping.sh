#!/bin/bash

INGRESS_IP=$(kubectl get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

for v in 1 2 3; do
  cmd="curl http://$INGRESS_IP/version -H \"version: v${v}\" -H \"Host: httpserver.sample.com\""
  echo $cmd
  eval $cmd
done
