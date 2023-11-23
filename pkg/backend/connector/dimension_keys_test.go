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

type clientMock struct {
	mock.Mock
}

// Returns a list of all available dimensions
func (clientmock *clientMock) ListDimensionKeys(ctx context.Context, in *v3.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v3.ListDimensionKeysResponse, error) {
	args := clientmock.Called(ctx, in, opts)
	if v, ok := args.Get(0).(*v3.ListDimensionKeysResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

// Returns a list of all dimension values for a certain dimension
func (clientmock *clientMock) ListDimensionValues(ctx context.Context, in *v3.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v3.ListDimensionValuesResponse, error) {
	panic("not implemented") // TODO: Implement
}

// Returns all metrics from the system
func (clientmock *clientMock) ListMetrics(ctx context.Context, in *v3.ListMetricsRequest, opts ...grpc.CallOption) (*v3.ListMetricsResponse, error) {
	panic("not implemented") // TODO: Implement
}

// Gets the options for the specified query type
func (clientmock *clientMock) GetQueryOptions(ctx context.Context, in *v3.GetOptionsRequest, opts ...grpc.CallOption) (*v3.GetOptionsResponse, error) {
	panic("not implemented") // TODO: Implement
}

// Gets the last known value for one or more metrics
func (clientmock *clientMock) GetMetricValue(ctx context.Context, in *v3.GetMetricValueRequest, opts ...grpc.CallOption) (*v3.GetMetricValueResponse, error) {
	panic("not implemented") // TODO: Implement
}

// Gets the history for one or more metrics
func (clientmock *clientMock) GetMetricHistory(ctx context.Context, in *v3.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v3.GetMetricHistoryResponse, error) {
	panic("not implemented") // TODO: Implement
}

// Gets the history for one or more metrics
func (clientmock *clientMock) GetMetricAggregate(ctx context.Context, in *v3.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v3.GetMetricAggregateResponse, error) {
	panic("not implemented") // TODO: Implement
}
func (clientmock *clientMock) Dispose() {
	panic("not implemented") // TODO: Implement
}

func TestListDimensionKeys(t *testing.T) {
	m := &clientMock{}
	req := models.GetDimensionKeysRequest{
		Filter: "filter",
		SelectedDimensions: []models.Dimension{
			{Key: "foo", Value: "bar"},
		},
	}
	m.On("ListDimensionKeys", mock.Anything, &v3.ListDimensionKeysRequest{
		Filter: req.Filter,
		SelectedDimensions: []*v3.Dimension{
			{Key: "foo", Value: "bar"},
		},
	}, mock.Anything).Return(&v3.ListDimensionKeysResponse{
		Results: []*v3.ListDimensionKeysResponse_Result{
			{
				Key:         "foo",
				Description: "bar",
			},
		},
	}, nil)
	res, err := ListDimensionKeys(context.TODO(), m, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	exp := &models.GetDimensionKeysResponse{
		Keys: []models.DimensionKeyDefinition{
			{Value: "foo", Label: "foo", Description: "bar"},
		},
	}
	assert.EqualValues(t, exp, res)
}
