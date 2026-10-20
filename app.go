package main

import (
	"context"
	"errors"
	"time"

	"github.com/kemalasliyuksek/bilge/internal/cleanup"
	"github.com/kemalasliyuksek/bilge/internal/engine"
	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

// callTimeout bounds every engine call so a hung daemon cannot freeze the UI.
const callTimeout = 30 * time.Second

// App holds the methods bound to the frontend. Removal is deliberately not
// bound yet: it will be exposed together with the preview screen.
type App struct {
	ctx    context.Context
	engine *engine.Engine
	// engineErr is kept instead of exiting so the UI can explain what is wrong
	// rather than the app failing to start.
	engineErr error
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.engine, a.engineErr = engine.New()
}

func (a *App) shutdown(context.Context) {
	if a.engine != nil {
		_ = a.engine.Close()
	}
}

// Version returns the build version shown in the UI.
func (a *App) Version() string {
	return version
}

// ScanResult is one snapshot of the engine and the cleanup report built from it.
type ScanResult struct {
	Inventory inventory.Inventory `json:"inventory"`
	Report    cleanup.Report      `json:"report"`
	SafeSize  int64               `json:"safeSize"`
}

func (a *App) Scan() (ScanResult, error) {
	ctx, cancel, err := a.call()
	if err != nil {
		return ScanResult{}, err
	}
	defer cancel()

	inv, err := a.engine.Snapshot(ctx)
	if err != nil {
		return ScanResult{}, err
	}
	inv.TakenAt = time.Now()
	report := cleanup.Classify(inv)
	return ScanResult{Inventory: inv, Report: report, SafeSize: report.SafeSize()}, nil
}

func (a *App) StartContainer(id string) error {
	ctx, cancel, err := a.call()
	if err != nil {
		return err
	}
	defer cancel()
	return a.engine.StartContainer(ctx, id)
}

func (a *App) StopContainer(id string) error {
	ctx, cancel, err := a.call()
	if err != nil {
		return err
	}
	defer cancel()
	return a.engine.StopContainer(ctx, id)
}

func (a *App) RestartContainer(id string) error {
	ctx, cancel, err := a.call()
	if err != nil {
		return err
	}
	defer cancel()
	return a.engine.RestartContainer(ctx, id)
}

// call prepares a bounded context for one engine call, or reports why the
// engine is unavailable.
func (a *App) call() (context.Context, context.CancelFunc, error) {
	if a.engineErr != nil {
		return nil, nil, a.engineErr
	}
	if a.engine == nil {
		return nil, nil, errors.New("engine not ready")
	}
	ctx, cancel := context.WithTimeout(a.ctx, callTimeout)
	return ctx, cancel, nil
}
