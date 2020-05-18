package collector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/topine/ibm-spectrum-exporter/monitoring"
	"github.com/topine/ibm-spectrum-exporter/spectrumservice"
)

func init() {
	registerCollector("switch", true, newSwitchCollector)
}

type switchCollector struct {
	ibmSpectrumClient spectrumservice.Client
	logger            *zap.SugaredLogger
	metrics           map[int]*prometheus.Desc
}

// newPoolCollector returns a new Collector Pools information
func newSwitchCollector(config monitoring.MetricsConfig, logger *zap.Logger,
	spectrumClient spectrumservice.Client) (Collector, error) {
	labelNameSwitch := []string{"name"}

	metrics := make(map[int]*prometheus.Desc)

	for _, metric := range config.Metrics.Switches {
		metrics[metric.MetricID] = prometheus.NewDesc(metric.PrometheusName, metric.PrometheusHelp, labelNameSwitch, nil)
	}

	return &switchCollector{
		ibmSpectrumClient: spectrumClient,
		logger:            logger.Sugar(),
		metrics:           metrics,
	}, nil
}

func (c *switchCollector) UpdateDescribe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
}

func (c *switchCollector) Update(ch chan<- prometheus.Metric) error {
	collectedMetrics, err := c.ibmSpectrumClient.CollectFromSwitch("switch")
	if err != nil {
		c.logger.Error("Error during authentication.", err)
		return nil
	}

	spectrumMetrics := collectedMetrics.Metrics

	for _, spectrumMetric := range spectrumMetrics {
		for _, switchMetric := range spectrumMetric.SwitchAggregatedMetrics {
			for x := 1; x <= len(switchMetric.Current); x++ {
				// get the latest available metric
				current := switchMetric.Current[len(switchMetric.Current)-x]
				if current.Y != nil {
					ch <- prometheus.NewMetricWithTimestamp(time.Unix(0, current.X*int64(time.Millisecond)),
						prometheus.MustNewConstMetric(c.metrics[switchMetric.MetricID],
							prometheus.GaugeValue, *current.Y, switchMetric.DeviceName))
					break
				}
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 1, "switch")
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, collectedMetrics.CollectionDuration, "switch")
	return nil
}
