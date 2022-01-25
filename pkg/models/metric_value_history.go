package models

import (
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type MetricHistoryQuery struct {
	MetricBaseQuery
}

func UnmarshalToMetricHistoryQuery(dq *backend.DataQuery) (*MetricHistoryQuery, error) {
	query := &MetricHistoryQuery{}
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

func (q MetricBaseQuery) FormatDisplayName(metricID string, labels []*pb.Label) string {
	if q.DisplayName == "" {
		return metricID
	}

	ctx := newContext(q, metricID, labels)

	s, err := parseDisplayNameExpr(ctx, q.DisplayName)
	if err != nil {
		return s
	}
	return s
}
