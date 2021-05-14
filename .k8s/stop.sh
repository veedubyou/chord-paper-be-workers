#!/bin/bash

# Convenience script to undeploy everything

kubectl delete svc/chord-be-workers-svc
kubectl delete deploy/chord-be-workers
watch kubectl get all
