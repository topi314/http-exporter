package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
	"golang.org/x/exp/slog"

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
	Global  GlobalConfig       `toml:"global"`
	Log     LogConfig          `toml:"log"`
	Server  ServerConfig       `toml:"server"`
	Configs []exporters.Config `toml:"configs"`
}

type GlobalConfig struct {
	ScrapeInterval xtime.Duration `toml:"scrape_interval"`
	ScrapeTimeout  xtime.Duration `toml:"scape_timeout"`
}

type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    string     `toml:"format"`
	AddSource bool       `toml:"add_source"`
}

type ServerConfig struct {
	ListenAddr string `toml:"listen_addr"`
	Endpoint   string `toml:"endpoint"`
}
