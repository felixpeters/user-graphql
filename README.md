# User GraphQL API

Experimenting with GraphQL in combination with Go, the binary from this app
simply runs a GraphQL API which serves user objects.

Users have two attributes for now:
* id: String
* username: String

## Setup

The app can be built and run using make and Docker:
* Run `make compile` in order to build the binary
* Run `make build-image` in order to build a Docker image with Alpine base
* Run `make run-docker` in order to build and run the Docker image
