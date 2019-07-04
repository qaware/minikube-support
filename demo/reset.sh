#!/usr/bin/env bash
../bin/minikube-support uninstall certManager --purge
../bin/minikube-support uninstall ingress-controller --purge
../bin/minikube-support uninstall coredns
kubectl delete -f deployment.yaml