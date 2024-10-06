package exporters

import (
	"context"
	"errors"
	"log/slog"

	"github.com/topi314/prometheus-collectors/internal/xtime"
)

var ErrExporterNotFound = errors.New("exporter not found")

var exporters = make(map[string]NewFunc)

func Register(name string, new NewFunc) {
	if _, ok := exporters[name]; ok {
		panic("exporter already registered")
	}
	exporters[name] = new
}

type Config struct {
	Name     string         `toml:"name"`
	Type     string         `toml:"type"`
	Interval xtime.Duration `toml:"interval"`
	Timeout  xtime.Duration `toml:"timeout"`
	Options  map[string]any `toml:"options"`
}

func New(cfg Config, logger *slog.Logger) (Exporter, error) {
	newFunc, ok := exporters[cfg.Type]
	if !ok {
		return nil, ErrExporterNotFound
	}
	return newFunc(cfg, logger)

}

type NewFunc func(cfg Config, logger *slog.Logger) (Exporter, error)

type Exporter interface {
	Collect(ctx context.Context)

	Close() error
}
