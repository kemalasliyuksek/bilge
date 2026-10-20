package main

import "context"

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

// App holds the methods bound to the frontend.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Version returns the build version shown in the UI.
func (a *App) Version() string {
	return version
}
