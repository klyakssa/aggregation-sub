.PHONY: build run test docker-up fmt

BINARY_NAME=aggregation-sub
BUILD_DIR=bin
DOCKER_IMAGE=aggregation-sub:latest
DOCKER_COMPOSE_FILE=docker-compose.yml
SWAG_CMD=$(shell go env GOPATH)/bin/swag

fmt:
	gofmt -s -w .
	goimports -l -w .

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/aggregator

run:
	go run ./cmd/aggregator/main.go

docker-up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-rebuild:
	docker-compose down
	docker rmi aggregation-sub-app
	docker-compose build --no-cache
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

swag-init:
	$(SWAG_CMD) fmt
	$(SWAG_CMD) init -g cmd/aggregator/main.go -o ./docs --outputTypes yaml,json --parseDependency --parseInternal --parseDepth 3

all: fmt test build
