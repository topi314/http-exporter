package main

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/topi314/prometheus-collectors/exporters"
)

func startExporters(ctx context.Context, cfg Config) {
	slog.DebugContext(ctx, "starting exporters")

	var wg sync.WaitGroup
	for i := range cfg.Configs {
		wg.Add(1)

		config := cfg.Configs[i]
		if config.Interval == 0 {
			config.Interval = cfg.Global.ScrapeInterval
		}
		if config.Timeout == 0 {
			config.Timeout = cfg.Global.ScrapeTimeout
		}
		go func() {
			defer wg.Done()
			logger := slog.With(
				slog.String("name", config.Name),
				slog.String("type", config.Type),
				slog.Duration("interval", time.Duration(config.Interval)),
				slog.Duration("timeout", time.Duration(config.Timeout)),
			)
			collect(ctx, logger, config)
		}()
	}

	wg.Wait()
}

func collect(ctx context.Context, logger *slog.Logger, cfg exporters.Config) {
	slog.DebugContext(ctx, "starting exporter", slog.String("name", cfg.Name))
	exporter, err := exporters.New(cfg, logger)
	if err != nil {
		if errors.Is(err, exporters.ErrExporterNotFound) {
			slog.ErrorContext(ctx, "exporter type not found", slog.String("type", cfg.Type))
			return
		}
		logger.ErrorContext(ctx, "failed to create exporter", slog.Any("err", err))
		return
	}
	defer func() {
		if closeErr := exporter.Close(); closeErr != nil {
			logger.ErrorContext(ctx, "failed to close exporter", slog.Any("err", closeErr))
		}
	}()

	timer := time.NewTicker(time.Duration(cfg.Interval))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			doCollect(ctx, exporter, cfg)
		}
	}
}

func doCollect(ctx context.Context, exporter exporters.Exporter, cfg exporters.Config) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Timeout))
	defer cancel()
	exporter.Collect(ctx)
}
