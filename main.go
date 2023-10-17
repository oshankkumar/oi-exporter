package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		symbol     string
		listenAddr string
	)
	flag.StringVar(&symbol, "symbol", "BANKNIFTY", "Symbol of stock")
	flag.StringVar(&listenAddr, "addr", ":8080", "listen address")
	flag.Parse()

	logger := slog.Default()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	lister := &NSEClient{
		Doer:    http.DefaultClient,
		BaseURL: "https://www.nseindia.com",
	}

	prometheus.MustRegister(
		NewOpenInterestCollector(ctx, "oi_exporter", symbol, logger, lister),
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error in starting exporter", "error", err)
		}
	}()

	logger.Info("started oi-expoerter", "addr", listenAddr, "symbol", symbol)
	wg.Wait()
}
