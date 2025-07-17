package models

import (
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const (
	QueryMetricValue     = "GetMetricValue"
	QueryMetricHistory   = "GetMetricHistory"
	QueryMetricAggregate = "GetMetricAggregate"
)

type Dimension struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Metric struct {
	MetricId string `json:"metricId"`
}

type OptionValue struct {
	Value string `json:"value,omitempty"`
	Label string `json:"label,omitempty"`
}

type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func (tr TimeRange) IsZero() bool {
	return tr.From.IsZero() && tr.To.IsZero()
}

type MetricBaseQuery struct {
	Dimensions    []Dimension            `json:"dimensions"`
	Metrics       []Metric               `json:"metrics,omitempty"`
	NextToken     string                 `json:"nextToken,omitempty"`
	DisplayName   string                 `json:"displayName,omitempty"`
	Interval      time.Duration          `json:"-"`
	MaxDataPoints int64                  `json:"-"`
	QueryType     string                 `json:"-"`
	Options       map[string]OptionValue `json:"queryOptions,omitempty"`
	TimeRange     TimeRange              `json:"range"`
}

func applyDataQueryBaseFields(dq *backend.DataQuery, base *MetricBaseQuery) {
	if base.TimeRange.IsZero() {
		base.TimeRange = TimeRange(dq.TimeRange)
	}
	base.Interval = dq.Interval
	base.MaxDataPoints = dq.MaxDataPoints
	base.QueryType = dq.QueryType
}
