

# IBM Spectrum Exporter
[![CircleCI](https://circleci.com/gh/topine/ibm-spectrum-exporter/tree/master.svg?style=shield)](https://circleci.com/gh/topine/ibm-spectrum-exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/topine/ibm-spectrum-exporter.svg?maxAge=604800)](https://hub.docker.com/r/topine/ibm-spectrum-exporter/)
[![Go Report Card](https://goreportcard.com/badge/github.com/topine/ibm-spectrum-exporter)](https://goreportcard.com/report/github.com/topine/ibm-spectrum-exporter)

IBM Spectrum Exporter is a Prometheus exporter to collect & export data from IBM Spectrum.

## Collectors

Currently we have 3 collectors :

### Storage Systems
Collecting performance metrics from the Storage Systems and the volumes.

### Switches
Collecting performance metrics from the Switches

### Pools
Collecting properties from the pool, e.g. total capacity.

## Metrics selection
The selection of metrics to collect from IBM Spectrum is done based on a configuration file.

It will translate an internal metric ID into a prometheus metric.

The list of metrics available with the API can be verified with the following example :

Storage System

https://<spectrum-hostname>:9569/srm/REST/api/v1/StorageSystems/<storage id>/Performance

Switches

https://<spectrum-hostname>:9569/srm/REST/api/v1/Switches/<switche id>/Performance


### Metric config file definition

The metric config file describes the translation between IBM Spectrum metrics and the metrics to expose to Prometheus.

It's a simple YAML file following the below structure :

```
metrics:
  
  storage_systems:                                               # Section for Storage System only.
    - ibm_spectrum_metric_id: 1029
      prometheus_name: storage_invalid_link_transmission_rate
      prometheus_help: The average number of times per second that an invalid transmission word was detected by the port while the link did not experience any signal or synchronization loss.

  storage_systems_and_volumes:                                  # Section for Storage System and Volumes
    - ibm_spectrum_metric_id: 803                               # Internal ID of the metric to be exposed
      prometheus_name: storage_avg_read_io_ops_per_second       # Prometheus metric name to be exported
      prometheus_help: Average number of read operations per second (both sequential and non-sequential, if applicable), for a particular component over a particular time interval.

    - ibm_spectrum_metric_id: 806
      prometheus_name: storage_avg_write_io_ops_per_second
      prometheus_help: Average number of write operations per second (both sequential and non-sequential, if applicable), for a particular component over a particular time interval.

  switches:                                                     # Section for Switches
    - ibm_spectrum_metric_id: 860                               # Internal ID of the metric to be exposed            
      prometheus_name: storage_switcher_avg_total_mb_per        # Prometheus metric name to be exported
      prometheus_help: Average number of mebibytes (2^20 bytes) transferred per second.

  pools:                                                        # Section for Pools
    properties:                                                 # In this case the Pool property will be exposed
      - property_name: Capacity                                 # Internal Property name
        prometheus_name: storage_usable_capacity_GiB            # Prometheus metric name to be exported
        prometheus_help: Usable Capacity

      - property_name: Available Pool Space
        prometheus_name: storage_free_capacity_GiB
        prometheus_help: Free Capacity
```

### Metrics ouput 

```
exporter-ip:9741/metrics


# HELP storage_avg_read_io_ops_per_second Average number of read operations per second (both sequential and non-sequential, if applicable), for a particular component over a particular time interval.
# TYPE storage_avg_read_io_ops_per_second gauge
storage_avg_read_io_ops_per_second{name="SVCXXXX",type="storageSystem"} 3691.71 1589834596000
storage_avg_read_io_ops_per_second{name="Volume=Name",storage_name="SVCXXXX",type="volume"} 14.7 1589834556000

# HELP storage_switcher_avg_total_mb_per Average number of mebibytes (2^20 bytes) transferred per second.
# TYPE storage_switcher_avg_total_mb_per gauge
storage_switcher_avg_total_mb_per{name="switchname"} 10746.11 1589834720000

# HELP storage_usable_capacity_GiB Usable Capacity
# TYPE storage_usable_capacity_GiB gauge
storage_usable_capacity_GiB{pool_name="pool-name",storage_system="v1234"} 2299.48


```

## IBM Spectrum API connection

To connect to the API an user is needed with read permissions.

## Installation

The tool can be installed from pre-built docker image or the binaries can be downloaded from the Github releases page.

### Installing the Docker Image

```
docker pull topine/ibm-spectrum-exporter:latest
```

## Usage

```
./ibm-spectrum-exporter --base-url=BASE-URL --user=USER --password=PASSWORD [<flags>]

Flags:
  -h, --help                                     Show context-sensitive help (also try --help-long and --help-man).
      --collector.pool                           Enable the pool collector (default: enabled).
      --collector.pool.filter=".*"               Enable the pool collectorvregex filter (default: .*).
      --collector.storage                        Enable the storage collector (default: enabled).
      --collector.storage.filter=".*"            Enable the storage collectorvregex filter (default: .*).
      --collector.switch                         Enable the switch collector (default: enabled).
      --collector.switch.filter=".*"             Enable the switch collectorvregex filter (default: .*).
      --listen-address=":9741"                   Address on which to expose metrics and web interface.
      --telemetry-path="/metrics"                Path under which to expose metrics.
      --metric-config-path="metrics_conf.yaml"   Metric configuration file absolute path
  -t, --base-url=BASE-URL                        IBM Spectrum base url
      --cache-metrics                            Cache metrics to avoid multiple calls
      --collection-interval="@every 5m"          Metrics Collection interval
  -u, --user=USER               IBM Spectrum username
  -p, --password=PASSWORD       IBM Spectrum username                          
```


## Certificate and trust management

For the moment all certificates are accepted, including the invalid ones.
