#!/bin/bash

set -ex

VERSION_NUMBER=$1

./push_docker.sh $VERSION_NUMBER
./deploy_kubernetes_config.sh

kubectl scale --replicas=0 deployment timer-server
kubectl scale --replicas=1 deployment timer-server

exit 0
