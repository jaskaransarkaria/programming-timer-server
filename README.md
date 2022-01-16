# Pair Programming Timer - Server

[![Build Status](https://travis-ci.com/jaskaransarkaria/programming-timer-server.svg?branch=master)](https://travis-ci.com/jaskaransarkaria/programming-timer-server)

---

## tl;dr

Keep time and turn order when you are pair programming, you can find the client code [here](https://github.com/jaskaransarkaria/programming-timer).

## Git and Deployment

1) Branch from `master`
2) Make your changes
4) Merge back into `master`
5) *Deploy* by tagging mater eg.`v1.0.0`

## Stack

  * `Go` - An open source programming language that makes it easy to build simple, reliable, and efficient software.
  * `Kubernetes` - An open-source system for automating deployment, scaling, and management of containerized applications.
  * `scripts/` - Build and deploy with bash scripts.

## Getting Started

Install `kubectl` from [here](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

Set up credentials to a kubernetes cluster.

To run the server locally

  * `go run main.go --addr localhost:8080`

## Useful Commands

  * `go run $ENTRY_FILE --addr $ADDR_AND_PORT_TO_SERVER_FROM`
  * `go build -o main .` - Build the Go code and output as "main"
  * `scripts/deploy.sh $VERSION_NUMBER` - Deploys any changes to kubernetes manifests, builds a new docker image, pushes it to docker hub and finally scales the deployment to pull the newly created image.
  * `scripts/deploy_kubernetes_config.sh` - Deploys just kubernetes manifest changes (kubernetes secret is excluded from the script).
  * `scripts/push_docker.sh $VERSION_NUMBER` - Builds and pushes the code to dockerhub with a $VERSION_NUMBER as a tag.
  * `go test -v ./...` - Runs all tests

## Deployment

Travis CI will run tests on each push and will deploy when master is tagged.

Local deployment is driven by bash scripts found in `scripts/`. _You must currently cd into scripts/ to execute them_.

To deploy your changes run (see "Useful Commands"):

  `./scripts/deploy.sh $VERSION_NUMBER`

> **NOTE** - If you change the VERSION_NUMBER of the docker image you must manually change the associated tag in `.kubernetes/deployment.yaml`. Use `scripts/deploy_kubernetes_config.sh` for updating just  k8 config.

### Todos

- [x] Add basic tests to cover Session
- [x] Add travis CI/ CD  & git branch rules/protection
- [x] Add in notifications and prompts to restart the timer

- [ ] Remove `func enableCors`
- [ ] Add environment config
- [ ] Tidy up bash scripts so can be called from proj root, prompt for required arguments and set VERSION_NUMBER so it is consistent across docker and k8 manifest.
