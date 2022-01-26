package models

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"time"
)

const (
	QueryMetricValue     = "GetMetricValue"
	QueryMetricHistory   = "GetMetricHistory"
	QueryMetricAggregate = "GetMetricAggregate"
	QueryDimensions      = "ListDimensionKeys"
	QueryDimensionValues = "ListDimensionValues"
	QueryMetrics         = "ListMetrics"
)

type Dimension struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Metric struct {
	MetricId string `json:"metricId"`
}

type MetricBaseQuery struct {
	Dimensions  []Dimension `json:"dimensions"`
	Metrics     []Metric    `json:"metrics,omitempty"`
	NextToken   string      `json:"nextToken,omitempty"`
	DisplayName string      `json:"displayName,omitempty"`

	Interval      time.Duration     `json:"-"`
	TimeRange     backend.TimeRange `json:"-"`
	MaxDataPoints int64             `json:"-"`
	QueryType     string            `json:"-"`
}
