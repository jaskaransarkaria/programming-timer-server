#!/bin/bash

set -ex

VERSION_NUMBER="$1"

docker build --no-cache -t jaskaransarkaria/timer-server:"$VERSION_NUMBER" ../
docker push jaskaransarkaria/timer-server:"$VERSION_NUMBER"

exit 0
