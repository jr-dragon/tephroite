package main

import (
	"context"
	"log/slog"
	"os"
)

var (
	// go build -ldflags "-X main.Name=tephroite -X main.Version=0.0.0"
	Name    string
	Version string
)

func main() {
	app := NewApp()

	ctx := context.TODO()
	if err := app.Run(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to run app", slog.Any("error", err))
		os.Exit(1)
	}
}
