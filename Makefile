.PHONY: build run test clean docker-build docker-up docker-down mocks

BINARY_NAME=homeclip-server
DOCKER_IMAGE=homeclip
GO_FILES=$(shell find . -name '*.go' -type f)

build:
	go build -o $(BINARY_NAME) ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

mocks:
	go generate ./...

docker-build:
	docker build -t $(DOCKER_IMAGE):latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down