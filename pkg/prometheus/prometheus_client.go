package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	API_PATH       = "/api/v1/"
	API_QUERY_PATH = "/api/v1/query"
	API_RANGE_PATH = "/api/v1/query_range"

	defaultTimeOut = time.Duration(60 * time.Second)
)

type MetricRestClient struct {
	client   *http.Client
	host     string
	username string
	password string
}

func NewRestClient(host string) (*MetricRestClient, error) {
	//1. get http client
	client := &http.Client{
		Timeout: defaultTimeOut,
	}

	//2. check whether it is using ssl
	addr, err := url.Parse(host)
	if err != nil {
		glog.Errorf("Invalid url:%v, %v", host, err)
		return nil, err
	}
	if addr.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	glog.V(2).Infof("Prometheus server address is: %v", host)

	return &MetricRestClient{
		client: client,
		host:   host,
	}, nil
}

func (c *MetricRestClient) SetUser(username, password string) {
	c.username = username
	c.password = password
}

func (c *MetricRestClient) Query(query string) (*promData, error) {
	p := fmt.Sprintf("%v%v", c.host, API_QUERY_PATH)
	glog.V(2).Infof("path=%v", p)

	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return nil, err
	}

	//1. set query
	q := req.URL.Query()
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	//2. set headers
	req.Header.Set("Accept", "application/json")
	if len(c.username) > 0 {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		glog.Errorf("Failed to send http request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Failed to read response: %v", err)
		return nil, err
	}

	var ss promeResponse
	if err := json.Unmarshal(result, &ss); err != nil {
		glog.Errorf("Failed to unmarshall respone: %v", err)
		return nil, err
	}

	if ss.Status == "error" {
		return nil, fmt.Errorf(ss.Error)
	}

	glog.V(3).Infof("resp: %++v", string(result))
	glog.V(3).Infof("metric: %+++v", ss)
	return ss.Data, nil
}

func (c *MetricRestClient) Test(query string) {
	//query := "rate(envoy_cluster_cds_internal_upstream_rq_200{job='envoy'}[3m])"
	//query := "envoy_cluster_sds_upstream_cx_length_ms"
	//query := "istio_request_count"
	result, err := c.Query(query)
	if err != nil {
		glog.Errorf("Failed to query: %v: %v", query, err)
		return
	}

	glog.V(2).Infof("query=%v, result.type=%v, \nresult.content=%v", query, result.ResultType, string(result.Result))

	var metrics []PrometheusMetric
	if err := json.Unmarshal(result.Result, &metrics); err != nil {
		glog.Errorf("Unmarshal failed: %v", err)
	}
	glog.V(2).Infof("dat = %++v", metrics)
}

func (c *MetricRestClient) GetMetrics(input PrometheusInput) ([]MetricData, error) {
	result := []MetricData{}

	//1. query
	qresult, err := c.Query(input.GetQuery())
	if err != nil {
		glog.Errorf("Failed to get metrics from prometheus: %v", err)
		return result, err
	}

	glog.V(4).Infof("result.type=%v, \n result: %+v",
		qresult.ResultType, string(qresult.Result))

	if qresult.ResultType != "vector" {
		err := fmt.Errorf("Unsupported result type: %v", qresult.ResultType)
		glog.Errorf(err.Error())
		return result, err
	}

	//2. parse/decode the value
	var resp []PrometheusMetric
	if err := json.Unmarshal(qresult.Result, &resp); err != nil {
		glog.Errorf("Failed to unmarshal: %v", err)
		return result, err
	}

	//3. assign the values
	for i := range resp {
		d, err := input.Parse(&(resp[i]))
		if err != nil {
			glog.Errorf("Pase value failed: %v", err)
			continue
		}

		result = append(result, d)
	}

	return result, nil
}

func (c *MetricRestClient) GetJobs() (string, error) {
	p := fmt.Sprintf("%v%v%v", c.host, API_PATH, "label/job/values")
	glog.V(2).Infof("path=%v", p)

	//1. prepare result
	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		glog.Errorf("Failed to send http request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	//2. read response
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Failed to read response: %v", err)
		return "", err
	}

	glog.V(3).Infof("resp: %++v", resp)
	glog.V(3).Infof("result: %++v", string(result))

	//3. parse response
	ss := promeResponse{}
	if err := json.Unmarshal(result, &ss); err != nil {
		glog.Errorf("Failed to unmarshall respone: %v", err)
	}

	if ss.Status == "error" {
		glog.Errorf("Error: %v", ss.Status)
	} else {
		glog.V(3).Infof("parsed: %++v", ss)
	}

	return string(result), nil
}
