package example

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"math"

	pclient "github.com/songbinliu/client_prometheus/pkg/prometheus"
)

// IstioQuery : generate queries for Istio-Prometheus metrics 
// qtype 0: pod.request-per-second
//       1: pod.latency
//       2: service.request-per-second
//       3: service.latency
type IstioQuery struct {
	qtype    int
	queryMap map[int]string
}

// IstioMetricData : hold the result of Istio-Prometheus data
type IstioMetricData struct {
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
	uuid   string
	dtype  int //0,1,2,3 same as qtype
}

// NewIstioQuery : create a new IstioQuery
func NewIstioQuery() *IstioQuery {
	q := &IstioQuery{
		qtype:    0,
		queryMap: make(map[int]string),
	}

	isPod := true
	q.queryMap[0] = getRPSExp(isPod)
	q.queryMap[1] = getLatencyExp(isPod)
	isPod = false
	q.queryMap[2] = getRPSExp(isPod)
	q.queryMap[3] = getLatencyExp(isPod)

	return q
}

func (q *IstioQuery) SetQueryType(t int) error {
	if t < 0 {
		err := fmt.Errorf("Invalid query type: %d, vs 0|1|2|3", t)
		glog.Error(err)
		return err
	}

	if t > len(q.queryMap) {
		err := fmt.Errorf("Invalid query type: %d, vs 0|1|2|3", t)
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

	labels := m.Labels
	if v, ok := labels["destination_uid"]; ok {
		d.Labels["destination_uid"] = v
	} else {
		return fmt.Errorf("No content for destination uid")
	}

	return nil
}

func (d *IstioMetricData) SetType(t int) {
	d.dtype = t
}

func (d *IstioMetricData) GetEntityID() (string, error) {
	label := "destination_uid"
	muid, ok := d.Labels[label]
	if !ok {
		err := fmt.Errorf("label-[%s] is missing", label)
		glog.Errorf(err.Error())
		return "", err
	}

	if d.dtype == 0 || d.dtype == 1 {
		return convertPodUID(muid)
	}

	if d.dtype == 2 || d.dtype == 3 {
		return convertSVCUID(muid)
	}

	err := fmt.Errorf("Invalid dtype: %v", d.dtype)
	glog.Error(err.Error())
	return "", err
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

	return buffer.String()
}
