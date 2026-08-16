.PHONY: all server clean

BIN_DIR := bin

all: server

server:
	@mkdir -p $(BIN_DIR)
	@go build -ldflags "-X main.Name=tephroite -X main.Version=0.0.0" -o $(BIN_DIR)/tp-server ./cmd/server

clean:
	rm -rf $(BIN_DIR)
