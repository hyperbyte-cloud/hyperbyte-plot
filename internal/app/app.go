package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"promviz/internal/backend"
	"promviz/internal/backend/influxdb"
	"promviz/internal/backend/influxdb1"
	"promviz/internal/backend/mock"
	"promviz/internal/backend/prom"
	"promviz/internal/config"
	"promviz/internal/ui"
)

// App represents the main application
type App struct {
	config       *config.Config
	backend      backend.Backend
	ui           *ui.TUI
	updateTicker *time.Ticker
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// New creates a new application instance
func New(configPath string) (*App, error) {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create backend (currently only Prometheus)
	backend, err := createBackend(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := backend.Connect(ctx); err != nil {
		return nil, err
	}

	// Create application context
	appCtx, appCancel := context.WithCancel(context.Background())

	app := &App{
		config:  cfg,
		backend: backend,
		ctx:     appCtx,
		cancel:  appCancel,
	}

	// Create UI with quit handler
	app.ui = ui.NewTUI(cfg.Queries, app.Stop)

	return app, nil
}

// createBackend creates the appropriate backend based on configuration
func createBackend(cfg *config.Config) (backend.Backend, error) {
	switch cfg.Backend {
	case "prometheus", "":
		promConfig := cfg.GetPrometheusConfig()
		return prom.NewClient(promConfig)
	case "influxdb":
		influxConfig := cfg.GetInfluxDBConfig()
		return influxdb.NewClient(influxConfig)
	case "influxdb1":
		influxConfig := cfg.GetInfluxDB1Config()
		return influxdb1.NewClient(influxConfig)
	case "mock":
		mockConfig := cfg.GetMockConfig()
		return mock.NewClient(mockConfig), nil
	default:
		return nil, fmt.Errorf("unsupported backend: %s (supported: prometheus, influxdb, influxdb1, mock)", cfg.Backend)
	}
}

// Start begins the application
func (a *App) Start() error {
	// Start periodic updates
	a.updateTicker = time.NewTicker(5 * time.Second)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.updateLoop()
	}()

	// Initial update
	go a.updateMetrics()

	// Start the TUI (this blocks until quit)
	return a.ui.Run()
}

// Stop gracefully shuts down the application
func (a *App) Stop() {
	if a.updateTicker != nil {
		a.updateTicker.Stop()
	}
	a.cancel()
	a.ui.Stop()

	// Wait for background goroutines to finish
	a.wg.Wait()

	// Close backend connection
	if a.backend != nil {
		a.backend.Close()
	}
}

// updateLoop runs the periodic metric updates
func (a *App) updateLoop() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-a.updateTicker.C:
			a.updateMetrics()
		}
	}
}

// updateMetrics fetches new data from the backend and updates the UI
func (a *App) updateMetrics() {
	ctx, cancel := context.WithTimeout(a.ctx, 3*time.Second)
	defer cancel()

	for i, query := range a.config.Queries {
		go func(idx int, q backend.Query) {
			timeSeries, err := a.backend.QueryTimeSeries(ctx, q.Expr)

			if err != nil {
				a.ui.UpdateTimeSeries(idx, nil, err)
				return
			}

			a.ui.UpdateTimeSeries(idx, timeSeries, nil)
		}(i, query)
	}
}
