APP := homeclip
BIN := ./bin/$(APP)
PORT ?= 8080
DATA := ./data

.PHONY: build run clean

build:
	go build -o $(BIN) ./cmd/homeclip

run: build
	DATA_DIR=$(DATA) PORT=$(PORT) $(BIN)

clean:
	rm -rf $(BIN) $(DATA)
