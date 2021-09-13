package models

import (
	"fmt"
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

func (d Dimension) String() string {
	return fmt.Sprintf("%s=%s", d.Key, d.Value)
}

type MetricBaseQuery struct {
	Dimensions []Dimension `json:"dimensions"`
	Metrics    []string    `json:"metrics,omitempty"`
	NextToken  string      `json:"nextToken,omitempty"`

	Interval      time.Duration     `json:"-"`
	TimeRange     backend.TimeRange `json:"-"`
	MaxDataPoints int64             `json:"-"`
	QueryType     string            `json:"-"`
}
