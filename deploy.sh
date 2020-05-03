#!/bin/bash

set -e

for FILENAME in .kubernetes/*.yaml; do
  [ -e "$FILENAME" ] || continue
  kubectl apply -f "$FILENAME"
done

exit 0
