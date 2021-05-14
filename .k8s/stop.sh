#!/bin/bash

# Convenience script to undeploy everything

kubectl delete svc/test-hooks-svc
kubectl delete deploy/test-hooks
watch kubectl get all
