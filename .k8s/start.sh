#!/bin/bash

# Convenience script to deploy everything

kubectl apply -f "deploy.yml" --record
watch kubectl get all
