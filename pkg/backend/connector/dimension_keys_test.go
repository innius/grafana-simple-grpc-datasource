package connector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
)

type clientMock struct {
	client.BackendAPIClient
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
