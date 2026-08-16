package main

import (
	"github.com/mazrean/kessoku"
)

//go:generate go tool kessoku $GOFILE
var _ = kessoku.Inject[*App](
	"NewApp",

	// app
	kessoku.Provide(func() *App {
		return &App{}
	}),
)
