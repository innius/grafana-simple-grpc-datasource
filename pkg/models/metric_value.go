package models

import (
	"encoding/json"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type MetricValueQuery struct {
	MetricBaseQuery
}

func UnmarshalToMetricValueQuery(dq *backend.DataQuery) (*MetricValueQuery, error) {
	query := &MetricValueQuery{}
	if err := json.Unmarshal(dq.JSON, query); err != nil {
		return nil, err
	}

	// add on the DataQuery params
	query.TimeRange = dq.TimeRange
	query.Interval = dq.Interval
	query.MaxDataPoints = dq.MaxDataPoints
	query.QueryType = dq.QueryType

	return query, nil
}

func (q MetricValueQuery) FormatDisplayName(metricID, value string) string {
	ctx := newContext(q.MetricBaseQuery, metricID)
	ctx["value"] = value

	s, err := parseDisplayNameExpr(ctx, q.DisplayName)
	if err != nil {
		return s
	}
	return s
}
