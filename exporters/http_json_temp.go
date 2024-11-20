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

const HTTPJSONTempType = "http-json-temp"

func init() {
	Register(HTTPJSONTempType, newHTTPJSONTemp)
}

func newHTTPJSONTemp(cfg Config, logger *slog.Logger) (Exporter, error) {
	var opts httpJSONOptions
	if err := xtoml.UnmarshalMap(cfg.Options, &opts); err != nil {
		return nil, fmt.Errorf("unmarshal http json temp options: %w", err)
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate http json temp options: %w", err)
	}

	temperature0 := getOrCreateGauge(prometheus.GaugeOpts{
		Name: opts.Metrics.Temperature0.Name,
		Help: opts.Metrics.Temperature0.Help,
	}, maps.Keys(opts.Metrics.Temperature0.Labels))

	temperature1 := getOrCreateGauge(prometheus.GaugeOpts{
		Name: opts.Metrics.Temperature1.Name,
		Help: opts.Metrics.Temperature1.Help,
	}, maps.Keys(opts.Metrics.Temperature1.Labels))

	return &httpJSONTempExporter{
		opts:   opts,
		logger: logger,
		gauges: httpJSONGauges{
			temperature0: temperature0,
			temperature1: temperature1,
		}, client: &http.Client{
			Timeout: time.Duration(cfg.Timeout),
		},
	}, nil
}

type httpJSONGauges struct {
	temperature0 *prometheus.GaugeVec
	temperature1 *prometheus.GaugeVec
}

type httpJSONTempExporter struct {
	opts   httpJSONOptions
	logger *slog.Logger
	gauges httpJSONGauges
	client *http.Client
}

func (e *httpJSONTempExporter) Collect(ctx context.Context) {
	e.logger.DebugContext(ctx, "collecting http-json-temp data")

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

	var data jsonData
	if err = json.NewDecoder(rs.Body).Decode(&data); err != nil {
		e.logger.Error("failed to decode response", slog.Any("err", err))
		return
	}

	e.gauges.temperature0.With(e.opts.Metrics.Temperature0.Labels).Set(data.Temperature0)
	e.gauges.temperature1.With(e.opts.Metrics.Temperature1.Labels).Set(data.Temperature1)
}

type jsonData struct {
	Temperature0 float64 `json:"temperature0"`
	Temperature1 float64 `json:"temperature1"`
}

func (e *httpJSONTempExporter) Close() error {
	e.logger.Debug("closing http-json-temp exporter")
	e.client.CloseIdleConnections()
	return nil
}

type httpJSONOptions struct {
	Metrics httpJSONMetricsConfig `toml:"metrics"`

	Address  string `toml:"address"`
	Insecure bool   `toml:"insecure"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

func (o httpJSONOptions) Validate() error {
	if o.Address == "" {
		return errors.New("address is required")
	}
	return nil
}

func (o httpJSONOptions) String() string {
	return fmt.Sprintf("\n address: %s\n insecure: %v\n username: %s\n password: %s\n metrics: %v",
		o.Address,
		o.Insecure,
		o.Username,
		strings.Repeat("*", len(o.Password)),
		o.Metrics,
	)
}

type httpJSONMetricsConfig struct {
	Temperature0 metricConfig `toml:"temperature0"`
	Temperature1 metricConfig `toml:"temperature1"`
}

func (c httpJSONMetricsConfig) Validate() error {
	if err := c.Temperature0.Validate(); err != nil {
		return fmt.Errorf("temperature0: %w", err)
	}
	if err := c.Temperature1.Validate(); err != nil {
		return fmt.Errorf("temperature1: %w", err)
	}
	return nil
}

func (c httpJSONMetricsConfig) String() string {
	return fmt.Sprintf("\n  temperature0: %s\n  temperature1: %s",
		c.Temperature0,
		c.Temperature1,
	)
}
