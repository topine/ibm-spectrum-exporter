package collector

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/topine/ibm-spectrum-exporter/monitoring"
	"github.com/topine/ibm-spectrum-exporter/spectrumservice"
)

var (
	svcInfo = prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "info"),
		"Storage System Info.",
		[]string{"type", "model", "name", "firmware", "ip_address"},
		nil,
	)
)

func init() {
	registerCollector("storage", true, newStorageCollector)
}

type storageCollector struct {
	ibmSpectrumClient spectrumservice.Client
	logger            *zap.SugaredLogger
	metrics           map[int]*prometheus.Desc
}

// newPoolCollector returns a new Collector Pools information
func newStorageCollector(config monitoring.MetricsConfig, logger *zap.Logger,
	spectrumClient spectrumservice.Client) (Collector, error) {
	labelNames := []string{"name", "type", "storage_name"}

	metrics := make(map[int]*prometheus.Desc)

	//transform the config into prometheus desc
	for _, metric := range config.Metrics.StorageSystems {
		metrics[metric.MetricID] = prometheus.NewDesc(metric.PrometheusName, metric.PrometheusHelp, labelNames, nil)
	}

	return &storageCollector{
		ibmSpectrumClient: spectrumClient,
		logger:            logger.Sugar(),
		metrics:           metrics,
	}, nil
}

func (c *storageCollector) UpdateDescribe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
	ch <- svcInfo
}

func (c *storageCollector) Update(ch chan<- prometheus.Metric) error {
	collectedMetrics, err := c.ibmSpectrumClient.CollectFromStorage("storage")
	if err != nil {
		c.logger.Error("Error during authentication.", err)
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 0)
		return err
	}

	spectrumMetrics := collectedMetrics.Metrics

	for _, spectrumMetric := range spectrumMetrics {
		for _, storageMetric := range spectrumMetric.StorageSystemMetrics {
			for x := 1; x <= len(storageMetric.Current); x++ {
				// get the latest available metric
				current := storageMetric.Current[len(storageMetric.Current)-x]
				if current.Y != nil {
					ch <- prometheus.NewMetricWithTimestamp(time.Unix(0, current.X*int64(time.Millisecond)),
						prometheus.MustNewConstMetric(c.metrics[storageMetric.MetricID],
							prometheus.GaugeValue, *current.Y, storageMetric.DeviceName, "storageSystem", ""))
					break
				}
			}
		}

		for _, volumeMetrics := range spectrumMetric.VolumeMetrics {
			for x := 1; x <= len(volumeMetrics.Current); x++ {
				// get the latest available metric
				current := volumeMetrics.Current[len(volumeMetrics.Current)-x]
				if current.Y != nil {
					ch <- prometheus.NewMetricWithTimestamp(time.Unix(0, current.X*int64(time.Millisecond)),
						prometheus.MustNewConstMetric(c.metrics[volumeMetrics.MetricID],
							prometheus.GaugeValue, *current.Y, strings.TrimSpace(volumeMetrics.DeviceName), "volume",
							strings.TrimSpace(volumeMetrics.ParentDeviceName)))
					break
				}
			}
		}
		ch <- prometheus.MustNewConstMetric(svcInfo, prometheus.GaugeValue, 1, spectrumMetric.Storage.Type,
			spectrumMetric.Storage.Model, spectrumMetric.Storage.Name, spectrumMetric.Storage.Firmware,
			spectrumMetric.Storage.IPAddress)
	}
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 1, "storage")
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, collectedMetrics.CollectionDuration, "storage")

	return nil
}
