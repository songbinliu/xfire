package example

import (
	"fmt"
	"github.com/golang/glog"
	pclient "github.com/songbinliu/client_prometheus/pkg/prometheus"
	"strings"
)

const (
	// NOTO: for istio 2.x, the prefix "istio_" should be removed
	TURBO_SVC_LATENCY_SUM   = "istio_turbo_service_latency_time_ms_sum"
	TURBO_SVC_LATENCY_COUNT = "istio_turbo_service_latency_time_ms_count"
	TURBO_SVC_REQUEST_COUNT = "istio_turbo_service_request_count"

	TURBO_POD_LATENCY_SUM   = "istio_turbo_pod_latency_time_ms_sum"
	TURBO_POD_LATENCY_COUNT = "istio_turbo_pod_latency_time_ms_count"
	TURBO_POD_REQUEST_COUNT = "istio_turbo_pod_request_count"

	TURBO_LATENCY_DURATION = "3m"

	K8S_PREFIX     = "kubernetes://"
	K8S_PREFIX_LEN = len(K8S_PREFIX)
)

func GetIstioMetric(client *pclient.MetricRestClient) {
	q := NewIstioQuery()

	for i := 0; i < 4; i++ {
		q.SetQueryType(i)
		result, err := client.GetMetrics(q)
		if err != nil {
			glog.Errorf("Failed to get metric: %v", err)
		}

		msg := "Pod QPS"
		if i == 1 {
			msg = "Pod Latency"
		} else if i == 2 {
			msg = "Service QPS"
		} else if i == 3 {
			msg = "Service Latency"
		}

		glog.V(2).Infof("====== %v =========", msg)
		for i := range result {
			glog.V(2).Infof("\t[%d] %v", i, result[i])
		}
	}
}

func getLatencyExp(pod bool) string {
	name_sum := ""
	name_count := ""
	if pod {
		name_sum = TURBO_POD_LATENCY_SUM
		name_count = TURBO_POD_LATENCY_COUNT
	} else {
		name_sum = TURBO_SVC_LATENCY_SUM
		name_count = TURBO_SVC_LATENCY_COUNT
	}
	du := TURBO_LATENCY_DURATION

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])/rate(%v{response_code=\"200\"}[%v])", name_sum, du, name_count, du)
	return result
}

// exp = rate(turbo_request_count{response_code="200",  source_service="unknown"}[3m])
func getRPSExp(pod bool) string {
	name_count := ""
	if pod {
		name_count = TURBO_POD_REQUEST_COUNT
	} else {
		name_count = TURBO_SVC_REQUEST_COUNT
	}
	du := TURBO_LATENCY_DURATION

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])", name_count, du)
	return result
}

// convert the UID from "kubernetes://<podName>.<namespace>" to "<namespace>/<podName>"
// for example, "kubernetes://video-671194421-vpxkh.default" to "default/video-671194421-vpxkh"
func convertPodUID(uid string) (string, error) {
	if !strings.HasPrefix(uid, K8S_PREFIX) {
		return "", fmt.Errorf("Not start with %v", K8S_PREFIX)
	}

	items := strings.Split(uid[K8S_PREFIX_LEN:], ".")
	if len(items) < 2 {
		return "", fmt.Errorf("Not enough fields: %v", uid[K8S_PREFIX_LEN:])
	}

	if len(items) > 2 {
		glog.Warningf("expected 2, got %d for: %v", len(items), uid[K8S_PREFIX_LEN:])
	}

	items[0] = strings.TrimSpace(items[0])
	items[1] = strings.TrimSpace(items[1])
	if len(items[0]) < 1 || len(items[1]) < 1 {
		return "", fmt.Errorf("Invalid fields: %v/%v", items[0], items[1])
	}

	nid := fmt.Sprintf("%s/%s", items[1], items[0])
	return nid, nil
}

// convert UID from "svcName.namespace.svc.cluster.local" to "svcName.namespace"
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
