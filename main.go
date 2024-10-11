package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version = "dev"
	Commit  = "unknown"
)

func main() {
	cfgPath := flag.String("config", "config.toml", "Path to config file")
	flag.Parse()

	slog.Info("Starting HTTP Exporter...", slog.String("version", Version), slog.String("commit", Commit), slog.String("config", *cfgPath))

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		slog.Error("Failed to load config", slog.Any("err", err))
		return
	}
	slog.Info("Loaded config", slog.String("config", cfg.String()))
	if err = cfg.Validate(); err != nil {
		slog.Error("Invalid config", slog.Any("err", err))
		return
	}

	setupLogger(cfg.Log)

	mux := http.NewServeMux()
	mux.Handle(cfg.Server.Endpoint, promhttp.Handler())
	mux.HandleFunc("/version", versionHandler(Version))
	server := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: mux,
	}
	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start server", slog.Any("err", err))
		}
	}()
	defer func() {
		if err = server.Shutdown(context.Background()); err != nil {
			slog.Error("Failed to shutdown server", slog.Any("err", err))
		}
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go startExporters(ctx, cfg)

	slog.Info("Started HTTP Exporter", slog.String("addr", cfg.Server.ListenAddr), slog.String("endpoint", cfg.Server.Endpoint))
	<-s
}

func setupLogger(cfg LogConfig) {
	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
	}
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}

func versionHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(version)); err != nil {
			slog.Error("Failed to write version response", slog.Any("err", err))
		}
	}
}
