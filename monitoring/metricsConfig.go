package monitoring

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// MetricsConfig : Struct to represent the config file
type MetricsConfig struct {
	Metrics struct {
		StorageSystems []struct {
			MetricID       int    `yaml:"ibm_spectrum_metric_id"`
			PrometheusName string `yaml:"prometheus_name"`
			PrometheusHelp string `yaml:"prometheus_help"`
		} `yaml:"storage_systems"`
		StorageSystemsAndVolumes []struct {
			MetricID       int    `yaml:"ibm_spectrum_metric_id"`
			PrometheusName string `yaml:"prometheus_name"`
			PrometheusHelp string `yaml:"prometheus_help"`
		} `yaml:"storage_systems_and_volumes"`
		Switches []struct {
			MetricID       int    `yaml:"ibm_spectrum_metric_id"`
			PrometheusName string `yaml:"prometheus_name"`
			PrometheusHelp string `yaml:"prometheus_help"`
		} `yaml:"switches"`
		Pools struct {
			Properties []struct {
				PropertyName   string `yaml:"property_name"`
				PrometheusName string `yaml:"prometheus_name"`
				PrometheusHelp string `yaml:"prometheus_help"`
			} `yaml:"properties"`
		} `yaml:"pools"`
	} `yaml:"metrics"`
}

// GetConf file from the given path
func (c *MetricsConfig) GetConf(filePath string) error {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, c)
	return err
}
