package api

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_valueQueryToInput(t *testing.T) {
	q := models.MetricValueQuery{
		MetricBaseQuery: models.MetricBaseQuery{
			Dimensions: []models.Dimension{
				{Key: "foo", Value: "bar"},
			},
			Metrics:       []string{"foo", "bar", "baz"},
			NextToken:     "",
			Interval:      0,
			TimeRange:     backend.TimeRange{},
			MaxDataPoints: 0,
			QueryType:     "",
		},
	}
	res := valueQueryToInput(q)

	require.Len(t, res.Dimensions, len(q.Dimensions))
	assert.Equal(t, q.Dimensions[0].Key, res.Dimensions[0].Key)
	assert.Equal(t, q.Dimensions[0].Value, res.Dimensions[0].Value)

	require.Len(t, res.Metrics, len(q.Metrics))
	assert.Equal(t, q.Metrics[0], res.Metrics[0].Id)
}
