package collector

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/topine/ibm-spectrum-exporter/monitoring"
	"github.com/topine/ibm-spectrum-exporter/spectrumservice"
)

func init() {
	registerCollector("pool", true, newPoolCollector)
}

type poolCollector struct {
	ibmSpectrumClient spectrumservice.Client
	logger            *zap.SugaredLogger
	properties        map[string]*prometheus.Desc
}

// newPoolCollector returns a new Collector Pools information
func newPoolCollector(config monitoring.MetricsConfig, logger *zap.Logger,
	spectrumClient spectrumservice.Client) (Collector, error) {

	labelPool := []string{"pool_name", "storage_system"}

	properties := make(map[string]*prometheus.Desc)

	for _, p := range config.Metrics.Pools.Properties {
		properties[p.PropertyName] = prometheus.NewDesc(p.PrometheusName, p.PrometheusHelp, labelPool, nil)
	}

	return &poolCollector{
		ibmSpectrumClient: spectrumClient,
		logger:            logger.Sugar(),
		properties:        properties,
	}, nil
}

func (c *poolCollector) UpdateDescribe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.properties {
		ch <- desc
	}
}

func (c *poolCollector) Update(ch chan<- prometheus.Metric) error {
	collectedMetrics, err := c.ibmSpectrumClient.CollectFromPools(*Filter["pool"])
	if err != nil || collectedMetrics == nil {
		c.logger.Error("Error getting Pools", err)
		return err
	}

	spectrumMetrics := collectedMetrics.Metrics

	for _, poolMetrics := range spectrumMetrics {
		p := poolMetrics.Pool
		t := reflect.TypeOf(p)
		v := reflect.ValueOf(p)

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if x, found := c.properties[f.Tag.Get("json")]; found && v.Field(i).String() != "" {
				value, err := strconv.ParseFloat(strings.ReplaceAll(v.Field(i).String(), ",", ""), 64)

				if err == nil {
					ch <- prometheus.MustNewConstMetric(x, prometheus.GaugeValue, value, p.Name, p.StorageSystem)
				} else {
					c.logger.Error("Error converting values.", err)
				}
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 1, "pool")
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, collectedMetrics.CollectionDuration, "pool")
	return nil
}
