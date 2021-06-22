package metrics

import "github.com/prometheus/client_golang/prometheus"

var registry *prometheus.Registry

func GetRegistry() *prometheus.Registry {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	return registry
}
