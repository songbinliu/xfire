package example

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"math"

	pclient "github.com/songbinliu/xfire/pkg/prometheus"
)

const (
	istioServiceLabel = "destination_service"

	metricSourceLabel = "metric_source"
	metricTypeLabel   = "metric_type"

	qpsValue     = "transaction_used"
	latencyValue = "latency_used"
)

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

// NewIstioQuery : create a new IstioQuery
func NewIstioQuery() *IstioQuery {
	q := &IstioQuery{
		qtype:    0,
		queryMap: make(map[int]string),
	}

	q.queryMap[0] = getRPSExp()
	q.queryMap[1] = getLatencyExp()

	return q
}

func (q *IstioQuery) SetQueryType(t int) error {
	if t < 0 || t > len(q.queryMap) {
		err := fmt.Errorf("Invalid query type: %d, vs 0|1", t)
		glog.Error(err)
		return err
	}

	q.qtype = t
	return nil
}

func (q *IstioQuery) GetQueryType() int {
	return q.qtype
}

func (q *IstioQuery) GetQuery() string {
	return q.queryMap[q.qtype]
}

func (q *IstioQuery) Parse(m *pclient.RawMetric) (pclient.MetricData, error) {
	d := NewIstioMetricData()
	d.SetType(q.qtype)
	if err := d.Parse(m); err != nil {
		glog.Errorf("Failed to parse metrics: %s", err)
		return nil, err
	}

	return d, nil
}

func (q *IstioQuery) String() string {
	var buffer bytes.Buffer

	for k, v := range q.queryMap {
		tmp := fmt.Sprintf("qtype:%d, query=%s", k, v)
		buffer.WriteString(tmp)
	}

	return buffer.String()
}

func NewIstioMetricData() *IstioMetricData {
	return &IstioMetricData{
		Labels: make(map[string]string),
	}
}

func (d *IstioMetricData) Parse(m *pclient.RawMetric) error {
	d.Value = float64(m.Value.Value)
	if math.IsNaN(d.Value) {
		return fmt.Errorf("Failed to convert value: NaN")
	}

	//1. select some original labels
	labels := m.Labels
	if v, ok := labels[istioServiceLabel]; ok {
		d.Labels[istioServiceLabel] = v
	} else {
		return fmt.Errorf("No content for destination uid")
	}

	//2. add other labes
	d.Labels[metricSourceLabel] = "istio"
	if d.dtype == 0 {
		d.Labels[metricTypeLabel] = qpsValue
	} else {
		d.Labels[metricTypeLabel] = latencyValue
	}

	return nil
}

func (d *IstioMetricData) SetType(t int) {
	d.dtype = t
}

func (d *IstioMetricData) GetEntityID() (string, error) {
	label := istioServiceLabel
	muid, ok := d.Labels[label]
	if !ok {
		err := fmt.Errorf("label-[%s] is missing", label)
		glog.Errorf(err.Error())
		return "", err
	}

	return convertSVCUID(muid)
}

func (d *IstioMetricData) GetValue() float64 {
	return d.Value
}

func (d *IstioMetricData) String() string {
	var buffer bytes.Buffer

	uid, err := d.GetEntityID()
	if err != nil {
		err := fmt.Errorf("Failed to get EntityID: %v", err)
		return err.Error()
	}

	content := fmt.Sprintf("uid=%v, value=%.5f", uid, d.GetValue())
	buffer.WriteString(content)
	content = fmt.Sprintf(" (%s, %s)", d.Labels[metricTypeLabel], d.Labels[metricSourceLabel])
	buffer.WriteString(content)

	return buffer.String()
}
