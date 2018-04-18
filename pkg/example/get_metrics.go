package example

import (
	"fmt"
	"github.com/golang/glog"
	pclient "github.com/songbinliu/xfire/pkg/prometheus"
	"strings"
)

const (
	// three metric names
	turboServiceLatencySumName = "istio_request_duration_sum"
	turboServiceLatencyCountName = "istio_request_duration_count"
	turboServiceRequestCountName = "istio_request_count"

	turboLatencyDuration = "3m"
)

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

func getLatencyExp() string {
	name_sum := turboServiceLatencySumName
	name_count := turboServiceLatencyCountName
	du := turboLatencyDuration

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])/rate(%v{response_code=\"200\"}[%v])", name_sum, du, name_count, du)
	return result
}

// exp = rate(turbo_request_count{response_code="200",  source_service="unknown"}[3m])
func getRPSExp() string {
	name_count := ""
	name_count = turboServiceRequestCountName
	du := turboLatencyDuration

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])", name_count, du)
	return result
}


// convert UID from "svcName.namespace.svc.cluster.local" to "svcName/namespace"
// for example, "productpage.default.svc.cluster.local" to "default/productpage"
func convertSVCUID(uid string) (string, error) {
	if uid == "unknown" {
		return "", fmt.Errorf("unknown")
	}

	//1. split it
	items := strings.Split(uid, ".")
	if len(items) < 3 {
		err := fmt.Errorf("Not enough fields %d Vs. 3", len(items))
		glog.V(3).Infof(err.Error())
		return "", err
	}

	//2. check the 3rd field
	items[0] = strings.TrimSpace(items[0])
	items[1] = strings.TrimSpace(items[1])
	items[2] = strings.TrimSpace(items[2])
	if items[2] != "svc" {
		err := fmt.Errorf("%v fields[2] should be [svc]: [%v]", uid, items[2])
		glog.V(3).Infof(err.Error())
		return "", err
	}

	//3. construct the new uid
	if len(items[0]) < 1 || len(items[1]) < 1 {
		err := fmt.Errorf("Invalid fields: %v/%v", items[0], items[1])
		glog.V(3).Infof(err.Error())
		return "", err
	}

	nid := fmt.Sprintf("%s/%s", items[1], items[0])
	return nid, nil
}
