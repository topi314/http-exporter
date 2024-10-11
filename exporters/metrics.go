package exporters

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var gaugeVecs = map[string]*prometheus.GaugeVec{}

func getOrCreateGauge(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	name := prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)

	gauge, ok := gaugeVecs[name]
	if !ok {
		gauge = promauto.NewGaugeVec(opts, labelNames)
		gaugeVecs[name] = gauge
	}

	return gauge
}

type metricConfig struct {
	Name   string            `toml:"name"`
	Help   string            `toml:"help"`
	Labels map[string]string `toml:"labels"`
}

func (c metricConfig) Validate() error {
	if c.Name == "" {
		return errors.New("metric config name is required")
	}
	return nil
}

func (c metricConfig) String() string {
	return fmt.Sprintf("\n  name: %s\n  help: %s\n  labels: %v",
		c.Name,
		c.Help,
		c.Labels,
	)
}
