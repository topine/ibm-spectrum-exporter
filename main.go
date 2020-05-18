package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/topine/ibm-spectrum-exporter/collector"
	"github.com/topine/ibm-spectrum-exporter/monitoring"
	"github.com/topine/ibm-spectrum-exporter/spectrumservice"
)

//BUILDTIME contains the build time
var BUILDTIME string

//TAG
var TAG string

//COMMIT
var COMMIT string

//VERSION
var VERSION string

func main() {

	var (
		//port allocation done in the prometheus https://github.com/prometheus/prometheus/wiki/Default-port-allocations
		addr               = kingpin.Flag("listen-address", "Address on which to expose metrics and web interface.").Default(":9741").String()
		metricsPath        = kingpin.Flag("telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		metricConfigPath   = kingpin.Flag("metric-config-path", "Metric configuration file absolute path").Default("metrics_conf.yaml").String()
		baseURL            = kingpin.Flag("base-url", "IBM Spectrum base url").Short('t').Required().String()
		cacheMetrics       = kingpin.Flag("cache-metrics", "Cache metrics to avoid multiple calls").Default("true").Bool()
		collectionInterval = kingpin.Flag("collection-interval", "Metrics Collection interval").Default("@every 5m").String()
		user               = kingpin.Flag("user", "IBM Spectrum username").Short('u').Required().String()
		password           = kingpin.Flag("password", "IBM Spectrum username").Short('p').Required().String()

		config         monitoring.MetricsConfig
		spectrumClient *spectrumservice.Client
		localCache     *cache.Cache
	)

	//kingpin.Version(version.Print("ibm-spectrum-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger, err := zap.NewDevelopment()
	//logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	if *baseURL == "" {
		logger.Sugar().Panic("IBM Spectrum base url missing")
	}

	err = config.GetConf(*metricConfigPath)
	if err != nil {
		logger.Sugar().Fatal("Error parsing the metrics configuration file: %v", err)
	}
	buildInfos()

	//starting cache
	localCache = cache.New(cache.NoExpiration, cache.NoExpiration)

	//set all the metrics
	spectrumClient = spectrumservice.NewClient(logger.Sugar(), config, localCache, *cacheMetrics,
		*user, *password, *baseURL)

	//Create a cron that will start a go routine to update the metrics
	c := cron.New()

	if *cacheMetrics {
		collectMetrics(logger, spectrumClient)
		//Create a cron that will start a go routine to update the metrics
		err = c.AddFunc(*collectionInterval, func() {
			collectMetrics(logger, spectrumClient)
		})
		if err != nil {
			logger.Sugar().Fatalf("Cannot schedule collection %v", err)
		} else {
			logger.Sugar().Infof("Scheduler started with success with interval %s \n", *collectionInterval)
		}
	}

	c.Start()

	spectrumCollector, err := collector.NewIbmSpectrumCollector(config, logger, *spectrumClient)
	if err != nil {
		logger.Sugar().Fatal("Error creating collector: %v", err)
	}

	prometheus.MustRegister(spectrumCollector)
	http.Handle(*metricsPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>IBM Spectrum Exporter</title></head>
			<body>
			<h1>IBM Spectrum Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))

		if err != nil {
			logger.Sugar().Panicf("Writing output page: %v", err)
		}
	})

	logger.Sugar().Infof("Exporter started with Success.")
	logger.Sugar().Fatal(http.ListenAndServe(*addr, nil))
}

func collectMetrics(logger *zap.Logger, spectrumClient *spectrumservice.Client) {
	logger.Sugar().Info("Starting to collect the metrics.")

	err := spectrumClient.CollectAndCacheMetrics(collector.Filter, collector.State)
	if err != nil {
		logger.Sugar().Errorf("error Collecting metrics for cache %v", err)
	}
	logger.Sugar().Info("Finished collecting metrics")
}

// buildInfos returns builds information
func buildInfos() {
	fmt.Println("Program started at: " + time.Now().String())
	fmt.Println("BUILDTIME=" + BUILDTIME)
	fmt.Println("TAG=" + TAG)
	fmt.Println("COMMIT=" + COMMIT)
	fmt.Println("VERSION=" + VERSION)
}
