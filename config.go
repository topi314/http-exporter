package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/topi314/prometheus-collectors/exporters"
	"github.com/topi314/prometheus-collectors/internal/xtime"
)

func defaultConfig() Config {
	return Config{
		Global: GlobalConfig{
			ScrapeInterval: xtime.Duration(1 * time.Minute),
			ScrapeTimeout:  xtime.Duration(10 * time.Second),
		},
		Log: LogConfig{
			Level:     slog.LevelInfo,
			Format:    "json",
			AddSource: false,
		},
		Server: ServerConfig{
			ListenAddr: ":2112",
			Endpoint:   "/metrics",
		},
	}
}

func loadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	cfg := defaultConfig()
	if err = toml.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file: %w", err)
	}
	return cfg, nil
}

type Config struct {
	Global  GlobalConfig      `toml:"global"`
	Log     LogConfig         `toml:"log"`
	Server  ServerConfig      `toml:"server"`
	Configs exporters.Configs `toml:"configs"`
}

func (c Config) Validate() error {
	var errs []error
	if err := c.Global.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("global: %w", err))
	}
	if err := c.Log.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("log: %w", err))
	}
	if err := c.Server.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("server: %w", err))
	}
	if err := c.Configs.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("configs: %w", err))
	}
	return errors.Join(errs...)
}

func (c Config) String() string {
	return fmt.Sprintf("\n global: %v\n log: %v\n server: %v\n configs: %v",
		c.Global,
		c.Log,
		c.Server,
		c.Configs,
	)
}

type GlobalConfig struct {
	ScrapeInterval xtime.Duration `toml:"scrape_interval"`
	ScrapeTimeout  xtime.Duration `toml:"scape_timeout"`
}

func (g GlobalConfig) Validate() error {
	var errs []error
	if g.ScrapeInterval <= 0 {
		errs = append(errs, fmt.Errorf("global config scrape_interval must be greater than 0"))
	}
	if g.ScrapeTimeout <= 0 {
		errs = append(errs, fmt.Errorf("global config scrape_timeout must be greater than 0"))
	}
	return errors.Join(errs...)
}

func (g GlobalConfig) String() string {
	return fmt.Sprintf("\n  scrape_interval: %s\n  scrape_timeout: %s",
		time.Duration(g.ScrapeInterval).String(),
		time.Duration(g.ScrapeTimeout).String(),
	)
}

type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    string     `toml:"format"`
	AddSource bool       `toml:"add_source"`
}

func (l LogConfig) Validate() error {
	if l.Format != "json" && l.Format != "text" {
		return fmt.Errorf("format must be json or text")
	}
	return nil
}

func (l LogConfig) String() string {
	return fmt.Sprintf("\n  level: %s\n  format: %s\n  add_source: %t",
		l.Level.String(),
		l.Format,
		l.AddSource,
	)
}

type ServerConfig struct {
	ListenAddr string `toml:"listen_addr"`
	Endpoint   string `toml:"endpoint"`
}

func (s ServerConfig) Validate() error {
	var errs []error
	if s.ListenAddr == "" {
		errs = append(errs, fmt.Errorf("server config listen_addr is required"))
	}
	if s.Endpoint == "" {
		errs = append(errs, fmt.Errorf("server config endpoint is required"))
	}
	return errors.Join(errs...)
}

func (s ServerConfig) String() string {
	return fmt.Sprintf("\n  listen_addr: %s\n  endpoint: %s",
		s.ListenAddr,
		s.Endpoint,
	)
}
