// Command server is the whole backend: the JSON API and the embedded
// frontend, in one binary.
//
// The daily season scheduler is gone. A season is issued on request rather
// than published once a day, so there is nothing to tick -- and with it went
// the goroutine, the UNIQUE (published_at) idempotency trick, and the whole
// question of leader election between instances.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/jvonmikael/lights-out/internal/api"
	"github.com/jvonmikael/lights-out/internal/sim"
	"github.com/jvonmikael/lights-out/internal/store"
	"github.com/jvonmikael/lights-out/internal/web"
)

type config struct {
	Port            string        `default:"8080"`
	DatabaseURL     string        `envconfig:"DATABASE_URL" required:"true"`
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"info"`
	SubmitPerMinute int           `envconfig:"SUBMIT_PER_MINUTE" default:"10"`
	ShutdownGrace   time.Duration `envconfig:"SHUTDOWN_GRACE" default:"15s"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return err
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		level = slog.LevelInfo
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	st, err := store.New(connectCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	if err := st.Migrate(); err != nil {
		return err
	}
	log.Info("migrations applied")

	var static http.Handler
	if web.Built() {
		static, err = web.Handler()
		if err != nil {
			return err
		}
	} else {
		log.Warn("no frontend build embedded; serving API only (run `make web`)")
	}

	srv := api.NewServer(api.Options{
		Store:           st,
		Logger:          log,
		Static:          static,
		SubmitPerMinute: cfg.SubmitPerMinute,
	})

	httpSrv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", httpSrv.Addr, "sim_version", sim.Version)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
		defer cancel()
		return httpSrv.Shutdown(shutdownCtx)
	}
}
