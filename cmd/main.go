package main

import (
	"flag"
	"github.com/golang/glog"

	"github.com/songbinliu/client_prometheus/pkg/example"
	"github.com/songbinliu/client_prometheus/pkg/prometheus"
)

var (
	prometheusHost string
	query          string
)

func parseFlags() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&prometheusHost, "promUrl", "http://localhost:19090", "the address of prometheus server")
	flag.StringVar(&query, "query", "http_requests_total", "the query for metrics from prometheus")
	flag.Parse()
}

func getJobs(mclient *prometheus.RestClient) {
	msg, err := mclient.GetJobs()
	if err != nil {
		glog.Errorf("Failed to get jobs: %v", err)
		return
	}
	glog.V(1).Infof("jobs: %v", msg)
}

func testPrometheus(mclient *prometheus.RestClient) {
	glog.V(2).Infof("Begin to test prometheus client...")
	getJobs(mclient)

	example.GetIstioMetric(mclient)
	return
}

func testGeneral(mclient *prometheus.RestClient) {
	glog.V(2).Infof("Begin to test general client ...")
	input := prometheus.NewGeneralPrometheusInput()

	query := "rate(istio_request_count[3m])"
	input.SetQuery(query)
	result, err := mclient.GetMetrics(input)
	if err != nil {
		glog.Errorf("Failed to get metrics for query: %v", input.GetQuery())
		return
	}
	for i := range result {
		glog.V(2).Infof("[%d] %++v", i, result[i])
	}
}

func main() {
	parseFlags()
	mclient, err := prometheus.NewRestClient(prometheusHost)
	if err != nil {
		glog.Fatalf("Failed to generate client: %v", err)
	}
	testPrometheus(mclient)
	testGeneral(mclient)

	return
}
