package prometheus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/model"
	"math"
)

// for internal use only
type promeResponse struct {
	Status    string    `json:"status"`
	Data      *promData `json:"data,omitempty"`
	ErrorType string    `json:"errorType,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type promData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

type RawMetric struct {
	Labels map[string]string `json:"metric"`
	Value  model.SamplePair  `json:"value"`
}

// interface to transfer the json.RawMessage to Value + Labels
type MetricData interface {
	GetEntityID() (string, error)
	GetValue() float64
}

type PrometheusInput interface {
	GetQuery() string
	Parse(metric *RawMetric) (MetricData, error)
}

// -----------------------------------------------------------
// an example implementation of PrometheusInput and MetricData
type GeneralMetricData struct {
	Labels map[string]string
	Value  float64
}

type GeneralPrometheusInput struct {
	query string
}

func NewGeneralPrometheusInput() *GeneralPrometheusInput {
	return &GeneralPrometheusInput{}
}

func (input *GeneralPrometheusInput) GetQuery() string {
	return input.query
}

func (input *GeneralPrometheusInput) SetQuery(q string) {
	input.query = q
}

func (input *GeneralPrometheusInput) Parse(m *RawMetric) (MetricData, error) {
	d := NewGeneralMetricData()

	for k, v := range m.Labels {
		d.Labels[k] = v
	}

	d.Value = float64(m.Value.Value)
	if math.IsNaN(d.Value) {
		return nil, fmt.Errorf("Failed to convert value: NaN")
	}
	return d, nil
}

func NewGeneralMetricData() *GeneralMetricData {
	return &GeneralMetricData{
		Labels: make(map[string]string),
	}
}

func (d *GeneralMetricData) GetEntityID() (string, error) {
	return "", nil
}

func (d *GeneralMetricData) GetValue() float64 {
	return d.Value
}

func (d *GeneralMetricData) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("value=%.6f\n", d.Value))
	for k, v := range d.Labels {
		buffer.WriteString(fmt.Sprintf("\t%v=%v\n", k, v))
	}
	return buffer.String()
}
