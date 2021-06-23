package check

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PiotrProkop/status/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// register metrics
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
	errLogger = log.New(os.Stdout, "ERROR:", log.Ldate|log.Ltime|log.Lshortfile)
)

const (
	up   float64 = 1
	down float64 = 0
)

// Doer is an interface that allows us to replace http.Client with any other struct implementing Do() function.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

var client Doer = &http.Client{}

// URL checks given url for status code and set appropiate Prometheus metric
func URL(url string) error {
	start := time.Now()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err

	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	// close response body and log errors
	defer func() {
		if err := response.Body.Close(); err != nil {
			errLogger.Println()
		}
	}()

	// we assume that only StatusCode == 200 means the service is UP
	if response.StatusCode == http.StatusOK {
		healthy.WithLabelValues(url).Set(up)
	} else {
		healthy.WithLabelValues(url).Set(down)
	}

	duration := time.Since(start).Milliseconds()

	responseTime.WithLabelValues(url).Set(float64(duration))

	return nil
}
