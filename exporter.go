package main

import (
	"context"
	"strconv"

	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

type OptionChainLister interface {
	ListOptionChain(ctx context.Context, symbol string) (*OptionChainIndex, error)
}

func NewOpenInterestCollector(ctx context.Context, namespace string, symbol string, logger *slog.Logger, optionChainLister OptionChainLister) *OpenInterestCollector {
	return &OpenInterestCollector{
		ctx:               ctx,
		symbol:            symbol,
		optionChainLister: optionChainLister,
		logger:            logger,

		// metrics
		openInterestM: prometheus.NewDesc(
			namespace+"_open_interest",
			"Total open interest",
			[]string{"option_type", "expiry_date", "strike_price"}, prometheus.Labels{"symbol": symbol},
		),
		lastPriceM: prometheus.NewDesc(
			namespace+"_last_price",
			"Last traded price",
			[]string{"option_type", "expiry_date", "strike_price"}, prometheus.Labels{"symbol": symbol},
		),
		underlyingValueM: prometheus.NewDesc(
			namespace+"_underlying_value",
			"Spot price of underlying asset",
			nil, prometheus.Labels{"symbol": symbol},
		),
		scrapeCountM: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scrape_count",
			Help:      "Total scrape count",
		}, []string{"symbol", "status"}),
	}
}

type OpenInterestCollector struct {
	symbol            string
	optionChainLister OptionChainLister
	ctx               context.Context
	logger            *slog.Logger

	//metrics
	openInterestM    *prometheus.Desc
	lastPriceM       *prometheus.Desc
	underlyingValueM *prometheus.Desc
	scrapeCountM     *prometheus.CounterVec
}

func (o *OpenInterestCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- o.lastPriceM
	ch <- o.openInterestM
	ch <- o.underlyingValueM
	o.scrapeCountM.Describe(ch)
}

func (e *OpenInterestCollector) Collect(ch chan<- prometheus.Metric) {
	status := "ok"
	if err := e.scrape(ch); err != nil {
		e.logger.Error("error in listing option chain data", "error", err)
		status = "fail"
	}

	e.scrapeCountM.WithLabelValues(e.symbol, status).Inc()
	e.scrapeCountM.Collect(ch)
}

func (e *OpenInterestCollector) scrape(ch chan<- prometheus.Metric) error {
	idx, err := e.optionChainLister.ListOptionChain(e.ctx, e.symbol)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(e.underlyingValueM, prometheus.GaugeValue, idx.Records.UnderlyingValue)

	for _, r := range idx.Records.Data {
		if r.PE.StrikePrice != 0 {
			lvs := []string{"PE", r.ExpiryDate, strconv.Itoa(r.PE.StrikePrice)}
			ch <- prometheus.MustNewConstMetric(e.openInterestM, prometheus.GaugeValue, r.PE.OpenInterest, lvs[:]...)
			ch <- prometheus.MustNewConstMetric(e.lastPriceM, prometheus.GaugeValue, r.PE.LastPrice, lvs[:]...)
		}
		if r.CE.StrikePrice != 0 {
			lvs := []string{"CE", r.ExpiryDate, strconv.Itoa(r.CE.StrikePrice)}
			ch <- prometheus.MustNewConstMetric(e.openInterestM, prometheus.GaugeValue, r.CE.OpenInterest, lvs[:]...)
			ch <- prometheus.MustNewConstMetric(e.lastPriceM, prometheus.GaugeValue, r.CE.LastPrice, lvs[:]...)
		}
	}

	return nil
}
