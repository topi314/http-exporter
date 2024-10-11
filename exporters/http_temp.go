package exporters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/maps"

	"github.com/topi314/prometheus-collectors/internal/xtoml"
)

const HTTPTempType = "http-temp"

func init() {
	Register(HTTPTempType, newHTTPTemp)
}

func newHTTPTemp(cfg Config, logger *slog.Logger) (Exporter, error) {
	var opts httpTempOptions
	if err := xtoml.UnmarshalMap(cfg.Options, &opts); err != nil {
		return nil, fmt.Errorf("unmarshal http temp options: %w", err)
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate http temp options: %w", err)
	}

	gauge := getOrCreateGauge(prometheus.GaugeOpts{
		Name: opts.Metric.Name,
		Help: opts.Metric.Help,
	}, maps.Keys(opts.Metric.Labels))

	return &httpTempExporter{
		opts:   opts,
		logger: logger,
		gauge:  gauge,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout),
		},
	}, nil
}

type httpTempExporter struct {
	opts   httpTempOptions
	logger *slog.Logger
	gauge  *prometheus.GaugeVec
	client *http.Client
}

func (e *httpTempExporter) Collect(ctx context.Context) {
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

	data, err := io.ReadAll(rs.Body)
	if err != nil {
		e.logger.Error("failed to read body", slog.Any("err", err))
		return
	}

	temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		e.logger.Error("failed to parse temperature", slog.Any("err", err))
		return
	}

	e.gauge.With(e.opts.Metric.Labels).Set(temp)
}

func (e *httpTempExporter) Close() error {
	e.logger.Debug("closing http-temp exporter")
	e.client.CloseIdleConnections()
	return nil
}

type httpTempOptions struct {
	Metric metricConfig `toml:"metric"`

	Address  string `toml:"address"`
	Insecure bool   `toml:"insecure"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

func (o httpTempOptions) Validate() error {
	var errs []error
	if o.Address == "" {
		errs = append(errs, fmt.Errorf("address is required"))
	}
	if err := o.Metric.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("metric: %w", err))
	}
	return errors.Join(errs...)
}

func (o httpTempOptions) String() string {
	return fmt.Sprintf("\n address: %s\n insecure: %t\n username: %s\n password: %s\n metric: %s",
		o.Address,
		o.Insecure,
		o.Username,
		strings.Repeat("*", len(o.Password)),
		o.Metric,
	)
}
