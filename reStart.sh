#!/bin/bash
operator-sdk generate k8s
operator-sdk generate crds
kubectl apply -f deploy/crds/app.kubelouis.com_mysqls_crd.yaml
operator-sdk run local

