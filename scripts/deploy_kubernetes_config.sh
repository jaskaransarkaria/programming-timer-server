#!/bin/bash

set -e

for FILENAME in ../.kubernetes/*.yaml; do
  [ -e "$FILENAME" ] || continue
  if [ "$FILENAME" == "../.kubernetes/secret.yaml" ]
  then
    continue
  fi
  kubectl apply -f "$FILENAME"
done

exit 0
