package models

import (
	"bytes"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"strings"
	"text/template"
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

func newContext(q MetricBaseQuery, metricID string) map[string]string {
	ctx := map[string]string{
		"metric": metricID,
	}
	for _, v := range q.Dimensions {
		ctx[v.Key] = v.Value
	}
	return ctx
}

func parseTemplate(alias string) (*template.Template, error) {
	t := template.New("alias")
	text := strings.Replace(alias, "{{", "{{.", -1)
	return t.Parse(text)
}

func parseDisplayNameExpr(ctx map[string]string, alias string) (string, error) {
	t, err := parseTemplate(alias)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = t.Execute(&b, ctx)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
