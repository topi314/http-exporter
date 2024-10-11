package exporters

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

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

type Configs []Config

func (c Configs) Validate() error {
	var errs []error
	for i := range c {
		if err := c[i].Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (c Configs) String() string {
	var str string
	for i := range c {
		str += c[i].String() + "\n"
	}
	return str
}

type Config struct {
	Name     string         `toml:"name"`
	Type     string         `toml:"type"`
	Interval xtime.Duration `toml:"interval"`
	Timeout  xtime.Duration `toml:"timeout"`
	Options  map[string]any `toml:"options"`
}

func (c Config) Validate() error {
	var errs []error
	if c.Name == "" {
		errs = append(errs, errors.New("exporter config name is required"))
	}
	if c.Type == "" {
		errs = append(errs, errors.New("exporter config type is required"))
	}
	if len(c.Options) == 0 {
		errs = append(errs, errors.New("exporter config options is required"))
	}
	return errors.Join(errs...)
}

func (c Config) String() string {
	return fmt.Sprintf("\n  name: %s\n  type: %s\n  interval: %s\n  timeout: %s\n  options: %v",
		c.Name,
		c.Type,
		time.Duration(c.Interval).String(),
		time.Duration(c.Timeout).String(),
		c.Options,
	)
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
