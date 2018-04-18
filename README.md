# client_prometheus
This is a [Prometheus](https://prometheus.io) metrics query sdk in Golang.

With this sdk, user can simply provide a query for a Prometheus server, the sdk will fetch the metrics,
and return a list of [MetricData](https://github.com/songbinliu/client_prometheus/blob/118d5ef7a0c31fe0a076587f97720b7cd55f50ff/pkg/prometheus/types.go#L31).

```go
// GetMetrics send a query to prometheus server, and return a list of MetricData
//   Note: it only support 'vector query: the data in the response is a 'vector'
//          not a 'matrix' (range query), 'string', or 'scalar'
//   (1) the RequestInput will generate a query;
//   (2) the RequestInput will parse the response into a list of MetricData

func (c *RestClient) GetMetrics(input RequestInput) ([]MetricData, error) {
...
}
```

The input of the sdk is [RequestInput](https://github.com/songbinliu/client_prometheus/blob/118d5ef7a0c31fe0a076587f97720b7cd55f50ff/pkg/prometheus/types.go#L36), which contains two things:
   * Query Generator 
   
        generate a query string, with which sdk will perform the query;
   * Result Parser
   
       parse the [RawMetric](https://github.com/songbinliu/client_prometheus/blob/118d5ef7a0c31fe0a076587f97720b7cd55f50ff/pkg/prometheus/types.go#L25) into user defined data structure;
       
# Run the examples
This SDK comes with some runnable examples (If you have a running prometheus)
```bash
sh script/build.sh
_output/metricClient --v=3
```

   
## the simple example
  The sdk already implemented a simple [BasicInput]() which will parse the `RawMetric` into 
  [BasicMetricData](https://github.com/songbinliu/client_prometheus/blob/118d5ef7a0c31fe0a076587f97720b7cd55f50ff/pkg/prometheus/types.go#L43)
  
  ```go
func testBasic(mclient *prometheus.RestClient) {
	glog.V(2).Infof("Begin to test basic client ...")
    query := "rate(http_requests_total[3m])"
	input := prometheus.NewBasicInput()

	input.SetQuery(query)
	result, err := mclient.GetMetrics(input)
	if err != nil {
		glog.Errorf("Failed to get metrics for query: %v", input.GetQuery())
		return
	}
	for i := range result {
		fmt.Printf("[%d] %++v\n", i, result[i])
	}
}
```

# the Istio example
In [this example](https://github.com/songbinliu/client_prometheus/tree/master/pkg/example), we can get service _latency_ and _request-per-second_ in the [Istio example](https://istio.io/docs/tasks/telemetry/metrics-logs.html). 
In addition, for each metric data, a uuid is generated based on the metric labels.

```go

// IstioQuery : generate queries for Istio-Prometheus metrics 
// qtype 0: svc.request-per-second
//       1: svc.latency
type IstioQuery struct {
	qtype    int
	queryMap map[int]string
}

// IstioMetricData : hold the result of Istio-Prometheus data
type IstioMetricData struct {
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
	uuid   string
	dtype  int //0,1 same as qtype
}


// GetIstioMetric get the 2 kinds of metrics from Istio-Prometheus
//    (1) service transaction-per-seconds;
//     (2) service latency;
func GetIstioMetric(client *pclient.RestClient) {
	q := NewIstioQuery()

	for i := 0; i < 2; i++ {
		q.SetQueryType(i)
		result, err := client.GetMetrics(q)
		if err != nil {
			glog.Errorf("Failed to get metric: %v", err)
		}

		msg := "Service Transaction per Seconds"
		if i == 1 {
			msg = "Service Latency"
		}

		glog.V(2).Infof("====== %v =========", msg)
		for i := range result {
			glog.V(2).Infof("\t[%d] %v", i, result[i])
		}
	}
}
```




