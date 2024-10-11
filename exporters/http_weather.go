package exporters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/maps"

	"github.com/topi314/prometheus-collectors/internal/xtoml"
)

const HTTPWeather = "http-weather"

func init() {
	Register(HTTPWeather, newHTTPWeather)
}

func newHTTPWeather(cfg Config, logger *slog.Logger) (Exporter, error) {
	var opts httpWeatherOptions
	if err := xtoml.UnmarshalMap(cfg.Options, &opts); err != nil {
		return nil, fmt.Errorf("unmarshal http weather options: %w", err)
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate http weather options: %w", err)
	}

	temperature0 := getOrCreateGauge(prometheus.GaugeOpts{
		Name: opts.Metrics.Temperature0.Name,
		Help: opts.Metrics.Temperature0.Help,
	}, maps.Keys(opts.Metrics.Temperature0.Labels))

	temperature1 := getOrCreateGauge(prometheus.GaugeOpts{
		Name: opts.Metrics.Temperature1.Name,
		Help: opts.Metrics.Temperature1.Help,
	}, maps.Keys(opts.Metrics.Temperature1.Labels))

	temperature2 := getOrCreateGauge(prometheus.GaugeOpts{
		Name: opts.Metrics.Temperature2.Name,
		Help: opts.Metrics.Temperature2.Help,
	}, maps.Keys(opts.Metrics.Temperature2.Labels))

	humidity := getOrCreateGauge(prometheus.GaugeOpts{
		Name: opts.Metrics.Humidity.Name,
		Help: opts.Metrics.Humidity.Help,
	}, maps.Keys(opts.Metrics.Humidity.Labels))

	pressure := getOrCreateGauge(prometheus.GaugeOpts{
		Name: opts.Metrics.Pressure.Name,
		Help: opts.Metrics.Pressure.Help,
	}, maps.Keys(opts.Metrics.Pressure.Labels))

	return &httpWeatherExporter{
		opts:   opts,
		logger: logger,
		gauges: httpWeatherGauges{
			temperature0: temperature0,
			temperature1: temperature1,
			temperature2: temperature2,
			humidity:     humidity,
			pressure:     pressure,
		},
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout),
		},
	}, nil
}

type httpWeatherGauges struct {
	temperature0 *prometheus.GaugeVec
	temperature1 *prometheus.GaugeVec
	temperature2 *prometheus.GaugeVec
	humidity     *prometheus.GaugeVec
	pressure     *prometheus.GaugeVec
}

type httpWeatherExporter struct {
	opts   httpWeatherOptions
	logger *slog.Logger
	gauges httpWeatherGauges
	client *http.Client
}

func (e *httpWeatherExporter) Collect(ctx context.Context) {
	e.logger.DebugContext(ctx, "collecting http-temp data")

	scheme := "https"
	if e.opts.Insecure {
		scheme = "http"
	}

	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s://%s", scheme, e.opts.Address), nil)
	if err != nil {
		e.logger.Error("failed to create request", slog.Any("err", err))
		return
	}

	if e.opts.Username != "" && e.opts.Password != "" {
		rq.SetBasicAuth(e.opts.Username, e.opts.Password)
	}

	rs, err := e.client.Do(rq)
	if err != nil {
		e.logger.Error("failed to do request", slog.Any("err", err))
		return
	}
	defer func() {
		if closeErr := rs.Body.Close(); closeErr != nil {
			e.logger.Error("failed to close body", slog.Any("err", closeErr))
		}
	}()

	if rs.StatusCode != http.StatusOK {
		e.logger.Error("unexpected status code", slog.Any("code", rs.StatusCode))
		return
	}

	var data weatherData
	if err = json.NewDecoder(rs.Body).Decode(&data); err != nil {
		e.logger.Error("failed to decode response", slog.Any("err", err))
		return
	}

	e.gauges.temperature0.With(e.opts.Metrics.Temperature0.Labels).Set(data.Temperature0)
	e.gauges.temperature1.With(e.opts.Metrics.Temperature1.Labels).Set(data.Temperature1)
	e.gauges.temperature2.With(e.opts.Metrics.Temperature2.Labels).Set(data.Temperature2)
	e.gauges.humidity.With(e.opts.Metrics.Humidity.Labels).Set(data.Humidity)
	e.gauges.pressure.With(e.opts.Metrics.Pressure.Labels).Set(data.Pressure)
}

type weatherData struct {
	Temperature0 float64 `json:"temperature0"`
	Temperature1 float64 `json:"temperature1"`
	Temperature2 float64 `json:"temperature2"`
	Humidity     float64 `json:"humidity"`
	Pressure     float64 `json:"pressure"`
}

func (e *httpWeatherExporter) Close() error {
	e.logger.Debug("closing http-weather exporter")
	e.client.CloseIdleConnections()
	return nil
}

type httpWeatherOptions struct {
	Metrics httpWeatherMetricsConfig `toml:"metrics"`

	Address  string `toml:"address"`
	Insecure bool   `toml:"insecure"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

func (o httpWeatherOptions) Validate() error {
	if o.Address == "" {
		return errors.New("address is required")
	}
	return nil
}

func (o httpWeatherOptions) String() string {
	return fmt.Sprintf("\n address: %s\n insecure: %v\n username: %s\n password: %s\n metrics: %v",
		o.Address,
		o.Insecure,
		o.Username,
		strings.Repeat("*", len(o.Password)),
		o.Metrics,
	)
}

type httpWeatherMetricsConfig struct {
	Temperature0 metricConfig `toml:"temperature0"`
	Temperature1 metricConfig `toml:"temperature1"`
	Temperature2 metricConfig `toml:"temperature2"`
	Humidity     metricConfig `toml:"humidity"`
	Pressure     metricConfig `toml:"pressure"`
}

func (c httpWeatherMetricsConfig) Validate() error {
	if err := c.Temperature0.Validate(); err != nil {
		return fmt.Errorf("temperature0: %w", err)
	}
	if err := c.Temperature1.Validate(); err != nil {
		return fmt.Errorf("temperature1: %w", err)
	}
	if err := c.Temperature2.Validate(); err != nil {
		return fmt.Errorf("temperature2: %w", err)
	}
	if err := c.Humidity.Validate(); err != nil {
		return fmt.Errorf("humidity: %w", err)
	}
	if err := c.Pressure.Validate(); err != nil {
		return fmt.Errorf("pressure: %w", err)
	}
	return nil
}

func (c httpWeatherMetricsConfig) String() string {
	return fmt.Sprintf("\n  temperature0: %s\n  temperature1: %s\n  temperature2: %s\n  humidity: %s\n  pressure: %s",
		c.Temperature0,
		c.Temperature1,
		c.Temperature2,
		c.Humidity,
		c.Pressure,
	)
}
