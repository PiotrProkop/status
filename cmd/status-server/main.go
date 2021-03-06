package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PiotrProkop/status/internal/check"
	"github.com/PiotrProkop/status/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	shutdownTimeout = 25
)

var (
	logger   = log.New(os.Stdout, "server-status", log.Ldate|log.Ltime|log.Lshortfile)
	interval string
	port     int
)

func SpawnChecker(closeChan chan struct{}, interval time.Duration, url string) {
	// populate metrics, if the interval is long we need to make sure we are serving metrics from the startup of application
	if err := check.URL(url); err != nil {
		logger.Println(err)
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-closeChan:
			return
		case <-ticker.C:
			if err := check.URL(url); err != nil {
				logger.Println(err)
			}
		}
	}
}

func main() {
	flag.StringVar(&interval, "interval", "1s", "interval between checking urls")
	flag.IntVar(&port, "port", 8000, "port to listen on")
	flag.Parse()

	registry := metrics.GetRegistry()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// prepare channels
	done := make(chan struct{})
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)

	checkInterval, err := time.ParseDuration(interval)
	if err != nil {
		logger.Fatal(err)
	}

	// spawn checkers
	go SpawnChecker(done, checkInterval, "https://httpstat.us/503")
	go SpawnChecker(done, checkInterval, "https://httpstat.us/200")

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-quit
		log.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		// shutdown gracefully to drain all current connections
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}

		close(done)
	}()

	log.Println("Server is ready to serve metrics at", fmt.Sprintf(":%d/metrics", port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", ":8080", err)
	}

	<-done
	log.Println("Server stopped")
}
