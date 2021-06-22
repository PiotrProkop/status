package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/PiotrProkop/status/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func init() {
	registry := metrics.GetRegistry()
	registry.MustRegister(responseTime)
	registry.MustRegister(healthy)
}

var (
	responseTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sample_external_url_response_ms",
		Help: "Response time in ms",
	},
		[]string{
			"url",
		},
	)
	healthy = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sample_external_url_up",
		Help: "Whether service is up or down",
	},
		[]string{
			"url",
		},
	)
	logger = log.Logger{}
	up     = 1
	down   = 0
)

// Doer is an interface that allows us to replace http.Client with any other struct implementing Do() function.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

var client Doer = &http.Client{}

func CheckURL(url string) error {
	start := time.Now()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err

	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			logger.Println(err)
		}
	}()

	if response.StatusCode == http.StatusOK {
		healthy.WithLabelValues(url).Set(float64(up))
	} else {
		healthy.WithLabelValues(url).Set(float64(down))
	}

	duration := time.Since(start).Milliseconds()

	responseTime.WithLabelValues(url).Set(float64(duration))

	return nil
}
