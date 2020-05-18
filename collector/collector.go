package collector

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/topine/ibm-spectrum-exporter/monitoring"
	"github.com/topine/ibm-spectrum-exporter/spectrumservice"
)

const namespace = "storage"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"ibm_spectrum_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)

	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"ibm_spectrum_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error

	// Describe metrics
	UpdateDescribe(ch chan<- *prometheus.Desc)
}

// IbmSpectrumCollector implements the prometheus.Collector interface.
type IbmSpectrumCollector struct {
	Collectors        map[string]Collector
	ibmSpectrumClient spectrumservice.Client
	logger            *zap.SugaredLogger
	regex             string
	metrics           map[int]*prometheus.Desc
	properties        map[string]*prometheus.Desc
}

var (
	factories = make(map[string]func(config monitoring.MetricsConfig, logger *zap.Logger,
		spectrumClient spectrumservice.Client) (Collector, error))
	State  = make(map[string]*bool)
	Filter = make(map[string]*string)
)

func registerCollector(collector string, isDefaultEnabled bool, factory func(config monitoring.MetricsConfig, logger *zap.Logger,
	spectrumClient spectrumservice.Client) (Collector, error)) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Bool()
	State[collector] = flag

	flagFilterName := fmt.Sprintf("collector.%s.filter", collector)
	flagFilterHelp := fmt.Sprintf("Enable the %s collectorvregex filter (default: %s).", collector, ".*")
	flagFilter := kingpin.Flag(flagFilterName, flagFilterHelp).Default(".*").String()
	Filter[collector] = flagFilter

	factories[collector] = factory
}

// NewIbmSpectrumCollector create new collector instance
func NewIbmSpectrumCollector(config monitoring.MetricsConfig, logger *zap.Logger,
	spectrumClient spectrumservice.Client) (*IbmSpectrumCollector, error) {

	collectors := make(map[string]Collector)
	for key, enabled := range State {
		if *enabled {
			collector, err := factories[key](config, logger, spectrumClient)
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
		}
	}

	return &IbmSpectrumCollector{Collectors: collectors, logger: logger.Sugar()}, nil
}

// Describe all metrics
func (c *IbmSpectrumCollector) Describe(ch chan<- *prometheus.Desc) {
	c.logger.Info("Starting IBM Spectrum collect.")

	wg := sync.WaitGroup{}
	wg.Add(len(c.Collectors))
	for name, ca := range c.Collectors {
		go func(name string, ca Collector) {
			ca.UpdateDescribe(ch)
			wg.Done()
		}(name, ca)
	}
	wg.Wait()

	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc

	for _, metric := range c.metrics {
		ch <- metric
	}

	for _, desc := range c.properties {
		ch <- desc
	}
}

// CollectFromStorage the metrics from IBM Spectrum
func (c *IbmSpectrumCollector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Info("Starting IBM Spectrum collect.")

	wg := sync.WaitGroup{}
	wg.Add(len(c.Collectors))
	for name, ca := range c.Collectors {
		go func(name string, ca Collector) {
			err := ca.Update(ch)
			if err != nil {
				c.logger.Error("Error collecting metrics.", err)
			}
			wg.Done()
		}(name, ca)
	}
	wg.Wait()
}
