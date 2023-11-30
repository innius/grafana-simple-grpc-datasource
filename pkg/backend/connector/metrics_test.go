package connector

import (
	"context"
	"testing"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// Returns all metrics from the system
func (clientmock *clientMock) ListMetrics(ctx context.Context, in *v3.ListMetricsRequest, opts ...grpc.CallOption) (*v3.ListMetricsResponse, error) {
	args := clientmock.Called(ctx, in, opts)
	if v, ok := args.Get(0).(*v3.ListMetricsResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestListMetrics(t *testing.T) {
	m := &clientMock{}
	req := models.GetMetricsRequest{
		Filter: "filter",
		Dimensions: []models.Dimension{
			{Key: "foo", Value: "bar"},
		},
	}
	m.On("ListMetrics", mock.Anything, &v3.ListMetricsRequest{
		Filter: req.Filter,
		Dimensions: []*v3.Dimension{
			{Key: "foo", Value: "bar"},
		},
	}, mock.Anything).Return(&v3.ListMetricsResponse{
		Metrics: []*v3.ListMetricsResponse_Metric{
			{
				Name:        "foo",
				Description: "bar",
			},
		},
	}, nil)
	res, err := ListMetrics(context.TODO(), m, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	exp := &models.GetMetricsResponse{
		Metrics: []models.MetricDefinition{
			{Value: "foo", Label: "foo", Description: "bar"},
		},
	}
	assert.EqualValues(t, exp, res)
}
