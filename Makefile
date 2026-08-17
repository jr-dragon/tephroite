.PHONY: all init server lint test clean

BIN_DIR := bin
GOIMPORTS_EXCLUDE := ./cmd/server/kessoku_band.go

all: server

init:
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install golang.org/x/tools/cmd/goimports@latest

server:
	@mkdir -p $(BIN_DIR)
	@go generate ./...
	@go build -ldflags "-X main.Name=tephroite -X main.Version=0.0.0" -o $(BIN_DIR)/tp-server ./cmd/server

lint:
	@find . -type f -name '*.go' ! -path '$(GOIMPORTS_EXCLUDE)' -print0 | \
		xargs -0 goimports -w -local github.com/jr-dragon/tephroite

test:
	@go generate ./...
	@govulncheck ./...
	@go test ./...
	@go vet ./...

clean:
	rm -rf $(BIN_DIR)
