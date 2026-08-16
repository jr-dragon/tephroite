.PHONY: all init server lint test clean

BIN_DIR := bin

all: server

init:
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install golang.org/x/tools/cmd/goimports@latest

server:
	@mkdir -p $(BIN_DIR)
	@go generate ./...
	@go build -ldflags "-X main.Name=tephroite -X main.Version=0.0.0" -o $(BIN_DIR)/tp-server ./cmd/server

lint:
	@goimports -w -local github.com/jr-dragon/tephroite .

test:
	@go generate ./...
	@govulncheck ./...
	@go test ./...

clean:
	rm -rf $(BIN_DIR)
