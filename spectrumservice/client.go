package spectrumservice

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/topine/ibm-spectrum-exporter/monitoring"
)

const (
	// docs https://www.ibm.com/support/knowledgecenter/SS5R93_5.3.6/com.ibm.spectrum.sc.doc/mgr_rest_api_retrieving_cli.html
	authenticate = "/srm/j_security_check"
	// Storage Systems
	listStorageSystems       = "/srm/REST/api/v1/StorageSystems"
	storageSystemPerformance = "/srm/REST/api/v1/StorageSystems/Performance"
	listVolumes              = "/srm/REST/api/v1/StorageSystems/{storageSystemID}/Volumes"
	volumesPerformance       = "/srm/REST/api/v1/StorageSystems/{storageSystemID}/Volumes/Performance"
	// Switches
	listSwitches      = "/srm/REST/api/v1/Switches"
	switchPerformance = "/srm/REST/api/v1/Switches/Performance"
	// Pools
	listPools = "/srm/REST/api/v1/Pools"
)

type Client struct {
	Sugar        *zap.SugaredLogger
	Config       monitoring.MetricsConfig
	Username     string
	Password     string
	BaseURL      string
	LocalCache   *cache.Cache
	CacheMetrics bool
	httpClient   *http.Client
}

func NewClient(sugar *zap.SugaredLogger, config monitoring.MetricsConfig, localCache *cache.Cache, cacheMetrics bool,
	usr, pwd, baseURL string) *Client {
	netTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint (for now mandatory as we will not have a valid certificate )
	}
	netClient := &http.Client{
		Timeout:   time.Second * 300,
		Transport: netTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Client{Sugar: sugar, Config: config, Username: usr, Password: pwd,
		BaseURL: baseURL, LocalCache: localCache, CacheMetrics: cacheMetrics, httpClient: netClient}
}

func (c *Client) CollectFromStorage(filter string) (*CollectedStorageMetrics, error) {
	if c.CacheMetrics {
		if x, found := c.LocalCache.Get("collectedMetrics"); found {
			return x.(*CollectedStorageMetrics), nil
		}
		return nil, errors.New("metrics not found in cache")
	}
	return c.CollectStorageMetrics(filter)
}

func (c *Client) CollectFromSwitch(filter string) (*CollectedSwitchMetrics, error) {
	if c.CacheMetrics {
		if x, found := c.LocalCache.Get("collectedSwitchMetrics"); found {
			return x.(*CollectedSwitchMetrics), nil
		}
		return nil, errors.New(" Switch metrics not found in cache")
	}
	return c.CollectSwitchMetrics(filter)
}

func (c *Client) CollectFromPools(filter string) (*CollectedPoolMetrics, error) {
	if c.CacheMetrics {
		if x, found := c.LocalCache.Get("collectedPoolMetrics"); found {
			return x.(*CollectedPoolMetrics), nil
		}
		return nil, errors.New(" Switch metrics not found in cache")
	}
	return c.CollectPools(filter)
}

func (c *Client) CollectAndCacheMetrics(filters map[string]*string, collectorsState map[string]*bool) error {
	var err error
	var wg sync.WaitGroup

	wg.Add(len(collectorsState))

	if *collectorsState["storage"] {
		go func() {
			defer wg.Done()
			collectedMetrics, errStorage := c.CollectStorageMetrics(*filters["storage"])
			if errStorage != nil {
				//c.LocalCache.Delete("collectedMetrics")
				c.Sugar.Error("Error Collecting metrics for cache.", errStorage)
				err = errStorage
			}
			c.LocalCache.Set("collectedMetrics", collectedMetrics, -1)
		}()
	}

	if *collectorsState["switch"] {
		go func() {
			defer wg.Done()
			collectedSwitchMetrics, errSwitches := c.CollectSwitchMetrics(*filters["switch"])
			if errSwitches != nil {
				//c.LocalCache.Delete("collectedMetrics")
				c.Sugar.Error("Error Collecting switches metrics for cache.", errSwitches)
				err = errSwitches
			}
			c.LocalCache.Set("collectedSwitchMetrics", collectedSwitchMetrics, -1)
		}()
	}

	if *collectorsState["pool"] {
		go func() {
			defer wg.Done()
			collectedPoolMetrics, errPool := c.CollectPools(*filters["pool"])
			if errPool != nil {
				//c.LocalCache.Delete("collectedMetrics")
				c.Sugar.Error("Error Collecting pool metrics for cache.", errPool)
				err = errPool
			}
			c.LocalCache.Set("collectedPoolMetrics", collectedPoolMetrics, -1)
		}()
	}
	wg.Wait()
	return err
}

func (c *Client) CollectStorageMetrics(filter string) (*CollectedStorageMetrics, error) {
	begin := time.Now()
	var response []*StorageMetrics //nolint prealloc
	cookies, err := c.authenticate()
	if err != nil {
		c.Sugar.Error("Error during authentication.", err)
		return nil, err
	}

	storages, err := c.listStorageSystems(cookies, filter)
	if err != nil {
		c.Sugar.Error("Error getting storage systems list.", err)
		return nil, err
	}

	paramsMap := make(map[string]string)

	// will get all the stats from the last 10 minutes
	// metrics seems to be delyed
	timeInMillis := time.Now().Add(time.Duration(-10)*time.Minute).UnixNano() / 1000000

	paramsMap["startTime"] = strconv.FormatInt(timeInMillis, 10)

	var buffer bytes.Buffer
	for _, metric := range c.Config.Metrics.StorageSystems {
		buffer.WriteString(strconv.Itoa(metric.MetricID))
		buffer.WriteString(",")
	}
	paramsMap["metrics"] = buffer.String()
	paramsMap["granularity"] = "sample"
	// Spectrum sometimes is not returning the values if asking for all storage at once
	for _, storage := range storages {
		storageID := storage.ID
		storageName := storage.Name

		paramsMap["ids"] = storageID

		storageMetrics, err := c.collectStorageSystemMetrics(cookies, storageID, paramsMap)
		if err != nil {
			c.Sugar.Error("Error retrieving storage system metrics.", err)
			continue
		}

		volumesMap, err := c.listVolumes(cookies, storageID)
		if err != nil {
			c.Sugar.Errorf("Error listing volumes for storage %s. %v", storageName,
				err)
			continue
		}

		delete(paramsMap, "ids")
		volumesMetrics, err := c.collectVolumeMetrics(cookies, storageID, paramsMap)
		if err != nil {
			c.Sugar.Errorf("Error collecting volumes metrics for storage %s. %v", storageName,
				err)
			continue
		}

		response = append(response,
			&StorageMetrics{
				Storage:              storage,
				VolumeMap:            volumesMap,
				StorageSystemMetrics: storageMetrics,
				VolumeMetrics:        volumesMetrics})
	}

	duration := time.Since(begin)

	return &CollectedStorageMetrics{Metrics: response, CollectionDuration: duration.Seconds()}, nil
}

func (c *Client) listStorageSystems(cookies []*http.Cookie, regex string) ([]StorageSystem, error) {
	// retrieve the system storage metrics ( already aggregated )
	responseStorage, err := c.doRequest("GET", c.BaseURL+listStorageSystems, nil, cookies, nil)
	if err != nil {
		c.Sugar.Error("Error requesting storage system metrics.", err)
		return nil, err
	}

	var storageSystems []StorageSystem
	err = json.Unmarshal(responseStorage, &storageSystems)
	if err != nil {
		c.Sugar.Error("Error reading storage system metrics response.", err)
		c.Sugar.Errorf("Response received: %s", string(responseStorage))
		return nil, err
	}

	c.Sugar.Infof("Number of Storage volumes retrieved: %d", len(storageSystems))

	var storages []StorageSystem

	c.Sugar.Infof("Selecting storage system with regex : %s", regex)
	for _, storageSystem := range storageSystems {
		matched, err := regexp.MatchString(regex, strings.ToUpper(storageSystem.Name))
		if err != nil {
			c.Sugar.Error("Error matching regex.", err)
			return nil, err
		}
		if matched {
			storages = append(storages, storageSystem)
		}
	}

	return storages, nil
}

func (c *Client) collectStorageSystemMetrics(cookies []*http.Cookie, storageSystemID string,
	paramMap map[string]string) ([]MetricValue, error) {
	storagePerm, err := c.doRequest("GET", strings.Replace(c.BaseURL+storageSystemPerformance,
		"{storageSystemID}", storageSystemID, -1), nil, cookies, paramMap)
	if err != nil {
		c.Sugar.Error("Error during storage system metrics call.", err)
		return nil, err
	}

	var objArray []*json.RawMessage
	err = json.Unmarshal(storagePerm, &objArray)
	if err != nil {
		c.Sugar.Error("Error reading storage system metrics response.", err)
		c.Sugar.Errorf("Response received: %s", string(storagePerm))
		return nil, err
	}

	c.Sugar.Infof("Metrics received for storageID %s : %d .", storageSystemID, len(objArray)-1)

	//objArray[0] are the metrics description
	var metricsValue []MetricValue
	for i := 1; i < len(objArray); i++ {
		var metricValue MetricValue
		err = json.Unmarshal(*objArray[i], &metricValue)
		if err != nil {
			c.Sugar.Error("Error mapping storage system metrics.", err)
			return nil, err
		}
		metricsValue = append(metricsValue, metricValue)
	}

	return metricsValue, nil
}

func (c *Client) listVolumes(cookies []*http.Cookie, storageSystemID string) (map[string]string, error) {
	//lookup for all volumes of the storage system
	volResponse, err := c.doRequest("GET",
		strings.Replace(c.BaseURL+listVolumes, "{storageSystemID}", storageSystemID, -1),
		nil, cookies, nil)
	if err != nil {
		return nil, err
	}

	data := Volumes{}
	err = json.Unmarshal(volResponse, &data)
	if err != nil {
		c.Sugar.Error("Error reading volume list response.", err)
		c.Sugar.Errorf("Response received: %s", string(volResponse))
		return nil, err
	}

	volumeMap := make(map[string]string)
	for _, volume := range data {
		volumeMap[volume.ID] = volume.VolumeUniqueID
	}

	c.Sugar.Infof("Volumes retrieved for Storage System %s : %d", storageSystemID, len(volumeMap))

	return volumeMap, nil
}

func (c *Client) collectVolumeMetrics(cookies []*http.Cookie, storageSystemID string,
	paramMap map[string]string) ([]MetricValue, error) {
	volumesPerm, err := c.doRequest("GET",
		strings.Replace(c.BaseURL+volumesPerformance, "{storageSystemID}", storageSystemID, -1),
		nil, cookies, paramMap)
	if err != nil {
		c.Sugar.Error("Error during volume metrics call.", err)
		return nil, err
	}

	var objArray []*json.RawMessage
	err = json.Unmarshal(volumesPerm, &objArray)
	if err != nil {
		c.Sugar.Error("Error reading volume metrics response.", err)
		c.Sugar.Errorf("Response received: %s", string(volumesPerm))
		return nil, err
	}

	var metricsValue []MetricValue
	for i := 1; i < len(objArray); i++ {
		var metricValue MetricValue
		err = json.Unmarshal(*objArray[i], &metricValue)
		if err != nil {
			c.Sugar.Error("Error mapping volume system metrics.", err)
			return nil, err
		}

		metricsValue = append(metricsValue, metricValue)
	}

	return metricsValue, nil
}

func (c *Client) CollectSwitchMetrics(filter string) (*CollectedSwitchMetrics, error) {
	begin := time.Now()
	var response []*SwitchMetrics //nolint prealloc
	cookies, err := c.authenticate()
	if err != nil {
		c.Sugar.Error("Error during authentication.", err)
		return nil, err
	}

	switches, err := c.listSwitches(cookies)
	if err != nil {
		c.Sugar.Error("Error getting switches list.", err)
		return nil, err
	}

	paramsMap := make(map[string]string)

	// will get all the stats from the last 20 minutes
	// metrics seems to be delyed
	timeInMillis := time.Now().Add(time.Duration(-20)*time.Minute).UnixNano() / 1000000

	paramsMap["startTime"] = strconv.FormatInt(timeInMillis, 10)

	var buffer bytes.Buffer
	for _, metric := range c.Config.Metrics.Switches {
		buffer.WriteString(strconv.Itoa(metric.MetricID))
		buffer.WriteString(",")
	}
	paramsMap["metrics"] = buffer.String()
	paramsMap["granularity"] = "sample"

	for _, s := range switches {

		matched, err := regexp.MatchString(filter, strings.ToUpper(s.Name))
		if err != nil {
			c.Sugar.Error("Error matching regex.", err)
			return nil, err
		}
		if !matched {
			continue
		}

		switchID := s.ID
		//switchName := s.Name

		paramsMap["ids"] = switchID

		switchMetrics, err := c.collectSwitchMetrics(cookies, switchID, paramsMap)
		if err != nil {
			c.Sugar.Error("Error retrieving siwtches metrics.", err)
			continue
		}

		if switchMetrics != nil {
			response = append(response,
				&SwitchMetrics{Switch: s,
					SwitchAggregatedMetrics: switchMetrics})
		}
	}

	duration := time.Since(begin)

	return &CollectedSwitchMetrics{Metrics: response, CollectionDuration: duration.Seconds()}, nil
}

func (c *Client) listSwitches(cookies []*http.Cookie) ([]Switch, error) {
	// retrieve the system storage metrics ( already aggregated )
	responseSwitches, err := c.doRequest("GET", c.BaseURL+listSwitches, nil, cookies, nil)
	if err != nil {
		c.Sugar.Error("Error listing switches.", err)
		return nil, err
	}

	var switches []Switch
	err = json.Unmarshal(responseSwitches, &switches)
	if err != nil {
		c.Sugar.Error("Error reading switches.", err)
		c.Sugar.Errorf("Response received: %s", string(responseSwitches))
		return nil, err
	}

	c.Sugar.Infof("Number of Switches retrieved: %d", len(switches))

	return switches, nil
}

func (c *Client) collectSwitchMetrics(cookies []*http.Cookie, switchID string,
	paramMap map[string]string) ([]MetricValue, error) {

	switchPerm, err := c.doRequest("GET", c.BaseURL+switchPerformance, nil, cookies, paramMap)
	if err != nil {
		c.Sugar.Error("Error during switch metrics call.", err)
		return nil, err
	}

	var objArray []*json.RawMessage
	err = json.Unmarshal(switchPerm, &objArray)
	if err != nil {
		c.Sugar.Error("Error reading switch metrics response.", err)
		c.Sugar.Errorf("Response received: %s", string(switchPerm))
		return nil, err
	}

	c.Sugar.Infof("Metrics received for switch %s : %d .", switchID, len(objArray)-1)

	//objArray[0] are the metrics description
	var metricsValue []MetricValue
	for i := 1; i < len(objArray); i++ {
		var metricValue MetricValue
		err = json.Unmarshal(*objArray[i], &metricValue)
		if err != nil {
			c.Sugar.Error("Error mapping switches metrics.", err)
			return nil, err
		}
		metricsValue = append(metricsValue, metricValue)
	}

	return metricsValue, nil
}

func (c *Client) CollectPools(filter string) (*CollectedPoolMetrics, error) {
	begin := time.Now()
	var response []*PoolsMetrics //nolint prealloc
	cookies, err := c.authenticate()
	if err != nil {
		c.Sugar.Error("Error during authentication.", err)
		return nil, err
	}

	pools, err := c.listPools(cookies)
	if err != nil {
		c.Sugar.Error("Error getting pool list.", err)
		return nil, err
	}

	// Spectrum sometimes is not returning the values if asking for all storage at once
	for _, p := range pools {

		matched, err := regexp.MatchString(filter, strings.ToUpper(p.Name))
		if err != nil {
			c.Sugar.Error("Error matching regex.", err)
			return nil, err
		}
		if matched {
			response = append(response, &PoolsMetrics{Pool: p})
		}
	}
	duration := time.Since(begin)
	return &CollectedPoolMetrics{Metrics: response, CollectionDuration: duration.Seconds()}, nil
}

func (c *Client) listPools(cookies []*http.Cookie) ([]Pool, error) {
	// retrieve the system storage metrics ( already aggregated )
	responseStorage, err := c.doRequest("GET", c.BaseURL+listPools, nil, cookies, nil)
	if err != nil {
		c.Sugar.Error("Error requesting pool metrics.", err)
		return nil, err
	}

	var pools []Pool
	err = json.Unmarshal(responseStorage, &pools)
	if err != nil {
		c.Sugar.Error("Error reading list pools response.", err)
		c.Sugar.Errorf("Response received: %s", string(responseStorage))
		return nil, err
	}

	c.Sugar.Infof("Number of Pools retrieved: %d", len(pools))

	return pools, nil
}

func (c *Client) authenticate() ([]*http.Cookie, error) {

	payload := url.Values{}

	payload.Set("j_username", c.Username)
	payload.Set("j_password", c.Password)

	req, err := http.NewRequest("POST", c.BaseURL+authenticate, strings.NewReader(payload.Encode()))
	if err != nil {
		c.Sugar.Error("Error creating request.", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		c.Sugar.Error("Error during authentication.", err)
		return nil, err
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Sugar.Error("Error reading response", err)
		return nil, err
	}

	return resp.Cookies(), nil
}

// doRequest : method to call execute the http request
func (c *Client) doRequest(method, url string, requestBody io.Reader, cookies []*http.Cookie, paramsMap map[string]string) ([]byte, error) { //nolint unparam

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	if len(paramsMap) > 0 {
		queryParams := req.URL.Query()
		for key, value := range paramsMap {
			queryParams.Add(key, value)
		}
		req.URL.RawQuery = queryParams.Encode()
	}

	resp, err := c.httpClient.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if http.StatusOK != resp.StatusCode {
		return nil, fmt.Errorf("%client", body)
	}

	return body, nil
}
