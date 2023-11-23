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

// Returns a list of all dimension values for a certain dimension
func (clientmock *clientMock) ListDimensionValues(ctx context.Context, in *v3.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v3.ListDimensionValuesResponse, error) {
	return &v3.ListDimensionValuesResponse{
		Results: []*v3.ListDimensionValuesResponse_Result{
			{Value: "foo", Description: "bar"},
		},
	}, nil
}

func TestListDimensionValues(t *testing.T) {
	m := &clientMock{}
	req := models.GetDimensionValueRequest{
		Filter:       "filter",
		DimensionKey: "foo",
		SelectedDimensions: []models.Dimension{
			{Key: "foo", Value: "bar"},
		},
	}
	m.On("ListDimensionValues", mock.Anything, &v3.ListDimensionValuesRequest{
		Filter:       req.Filter,
		DimensionKey: "foo",
		SelectedDimensions: []*v3.Dimension{
			{Key: "foo", Value: "bar"},
		},
	}, mock.Anything).Return(&v3.ListDimensionValuesResponse{
		Results: []*v3.ListDimensionValuesResponse_Result{
			{
				Value:       "foo",
				Description: "bar",
			},
		},
	}, nil)
	res, err := ListDimensionValues(context.TODO(), m, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	exp := &models.GetDimensionValueResponse{
		Values: []models.DimensionValueDefinition{
			{Value: "foo", Label: "foo", Description: "bar"},
		},
	}
	assert.EqualValues(t, exp, res)
}
