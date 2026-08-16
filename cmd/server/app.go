package main

import "context"

type App struct{}

func (app *App) Run(_ context.Context) error {
	return nil
}
