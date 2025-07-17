package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalToMetricValueQuery(t *testing.T) {
	jsonData := `{"someField": "someValue"}`
	dq := &backend.DataQuery{
		JSON:          json.RawMessage(jsonData),
		Interval:      time.Second,
		MaxDataPoints: 100,
		QueryType:     "query",
		TimeRange: backend.TimeRange{
			From: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	query, err := UnmarshalToMetricValueQuery(dq)
	assert.NoError(t, err)
	assert.Equal(t, time.Second, query.Interval)
	assert.Equal(t, int64(100), query.MaxDataPoints)
	assert.Equal(t, "query", query.QueryType)
	assert.Equal(t, TimeRange(dq.TimeRange), query.TimeRange)
}

func TestApplyDataQueryBaseFields_SkipsNonZeroTimeRange(t *testing.T) {
	base := MetricBaseQuery{
		TimeRange: TimeRange{
			From: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}
	dq := &backend.DataQuery{
		Interval:      time.Second,
		MaxDataPoints: 50,
		QueryType:     "agg",
		TimeRange: backend.TimeRange{
			From: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	applyDataQueryBaseFields(dq, &base)

	// TimeRange should remain unchanged
	assert.Equal(t, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC), base.TimeRange.From)
	assert.Equal(t, time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC), base.TimeRange.To)
	// Other fields updated
	assert.Equal(t, time.Second, base.Interval)
	assert.Equal(t, int64(50), base.MaxDataPoints)
	assert.Equal(t, "agg", base.QueryType)
}
