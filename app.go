package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/kemalasliyuksek/bilge/internal/cleanup"
	"github.com/kemalasliyuksek/bilge/internal/engine"
	"github.com/kemalasliyuksek/bilge/internal/inventory"
	"github.com/kemalasliyuksek/bilge/internal/sweep"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

// callTimeout bounds every engine call so a hung daemon cannot freeze the UI.
// Removal gets longer: pruning a large build cache takes minutes.
const (
	callTimeout   = 30 * time.Second
	removeTimeout = 10 * time.Minute
)

// ProgressEvent carries one sweep.Result to the frontend while a removal runs.
const ProgressEvent = "sweep:progress"

// App holds the methods bound to the frontend.
type App struct {
	ctx    context.Context
	engine *engine.Engine
	// removing allows one removal at a time; a second click while one runs
	// must not start a parallel sweep over the same resources.
	removing sync.Mutex
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

// Remove deletes the confirmed items. The sweep package re-checks every item
// against a fresh scan, so a stale or tampered request cannot remove anything
// that is in use or that needs a separate confirmation.
func (a *App) Remove(reqs []sweep.Request) ([]sweep.Result, error) {
	if !a.removing.TryLock() {
		return nil, errors.New("a removal is already running")
	}
	defer a.removing.Unlock()

	if a.engineErr != nil {
		return nil, a.engineErr
	}
	if a.engine == nil {
		return nil, errors.New("engine not ready")
	}
	ctx, cancel := context.WithTimeout(a.ctx, removeTimeout)
	defer cancel()
	return sweep.New(a.engine).Run(ctx, reqs, func(r sweep.Result) {
		runtime.EventsEmit(a.ctx, ProgressEvent, r)
	})
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
