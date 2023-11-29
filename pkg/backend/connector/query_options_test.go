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

// Gets the options for the specified query type
func (clientmock *clientMock) GetQueryOptions(
	ctx context.Context,
	in *v3.GetOptionsRequest,
	opts ...grpc.CallOption,
) (*v3.GetOptionsResponse, error) {
	args := clientmock.Called(ctx, in, opts)
	if v, ok := args.Get(0).(*v3.GetOptionsResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestGetQueryOptions(t *testing.T) {
	m := &clientMock{}
	req := models.GetQueryOptionsRequest{
		QueryType: "GetMetricAggregate",
		SelectedOptions: map[string]string{
			"foo": "bar",
		},
	}
	m.On("GetQueryOptions", mock.Anything, &v3.GetOptionsRequest{
		QueryType: v3.GetOptionsRequest_GetMetricAggregate,
		SelectedOptions: map[string]string{
			"foo": "bar",
		},
	}, mock.Anything).Return(&v3.GetOptionsResponse{
		Options: []*v3.Option{
			{
				Id:          "foo",
				Type:        v3.Option_Enum,
				Description: "the foo option",
				Required:    true,
				Label:       "foo label",
				EnumValues: []*v3.EnumValue{
					{
						Id:          "bar",
						Label:       "bar label",
						Description: "bar description",
						Default:     true,
					},
				},
			},
		},
	}, nil)
	res, err := GetQueryOptionDefinitions(context.TODO(), m, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	exp := &models.GetQueryOptionsResponse{
		Options: models.Options{
			{
				ID:          "foo",
				Type:        "Enum",
				Description: "the foo option",
				Required:    true,
				Label:       "foo label",
				EnumValues: []models.EnumValue{
					{
						ID:          "bar",
						Label:       "bar label",
						Description: "bar description",
						Default:     true,
					},
				},
			},
		},
	}
	assert.EqualValues(t, exp, res)
}
