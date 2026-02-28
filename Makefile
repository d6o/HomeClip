APP := homeclip
BIN := ./bin/$(APP)
PORT ?= 8080
DATA := ./data

.PHONY: build run clean test

build:
	go build -o $(BIN) ./cmd/homeclip

run: build
	DATA_DIR=$(DATA) PORT=$(PORT) $(BIN)

test:
	go test ./...

clean:
	rm -rf $(BIN) $(DATA)
