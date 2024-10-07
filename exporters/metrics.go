package exporters

import (
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
