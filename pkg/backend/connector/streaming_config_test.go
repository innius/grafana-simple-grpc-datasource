package connector

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	v4 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v4"
)

func (m *mockBackendClient) GetStreamingQueryConfiguration(ctx context.Context, req *v4.GetStreamingQueryConfigurationRequest, opts ...grpc.CallOption) (*v4.GetStreamingQueryConfigurationResponse, error) {
	args := m.Called(ctx, req)
	rsp := args.Get(0)
	if rsp == nil {
		return nil, args.Error(1)
	}
	return rsp.(*v4.GetStreamingQueryConfigurationResponse), args.Error(1)
}

func TestGetStreamingQueryConfiguration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		query         models.StreamingQueryConfigurationRequest
		mockResp      *v4.GetStreamingQueryConfigurationResponse
		mockErr       error
		expectResp    *models.StreamingQueryConfigurationResponse
		expectErr     bool
		expectErrText string
	}{
		{
			name: "MetricValueQuery returns config",
			query: models.StreamingQueryConfigurationRequest{
				Query: models.MetricValueQuery{},
			},
			mockResp: &v4.GetStreamingQueryConfigurationResponse{
				LookBackPeriodLimit: 123,
				LoopInterval:        456,
			},
			expectResp: &models.StreamingQueryConfigurationResponse{
				LookBackPeriodLimit: 123 * time.Millisecond,
				LoopInterval:        456 * time.Millisecond,
			},
		},
		{
			name: "MetricHistoryQuery returns config",
			query: models.StreamingQueryConfigurationRequest{
				Query: models.MetricHistoryQuery{},
			},
			mockResp: &v4.GetStreamingQueryConfigurationResponse{
				LookBackPeriodLimit: 789,
				LoopInterval:        1011,
			},
			expectResp: &models.StreamingQueryConfigurationResponse{
				LookBackPeriodLimit: 789 * time.Millisecond,
				LoopInterval:        1011 * time.Millisecond,
			},
		},
		{
			name: "MetricAggregateQuery returns config",
			query: models.StreamingQueryConfigurationRequest{
				Query: models.MetricAggregateQuery{},
			},
			mockResp: &v4.GetStreamingQueryConfigurationResponse{
				LookBackPeriodLimit: 222,
				LoopInterval:        333,
			},
			expectResp: &models.StreamingQueryConfigurationResponse{
				LookBackPeriodLimit: 222 * time.Millisecond,
				LoopInterval:        333 * time.Millisecond,
			},
		},
		{
			name: "Unsupported query type returns error",
			query: models.StreamingQueryConfigurationRequest{
				Query: "unsupported",
			},
			expectErr:     true,
			expectErrText: "unsupported query type for streaming configuration",
		},
		{
			name: "Backend client error is wrapped",
			query: models.StreamingQueryConfigurationRequest{
				Query: models.MetricValueQuery{},
			},
			mockErr:       errors.New("backend error"),
			expectErr:     true,
			expectErrText: "failed to get streaming query configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBackendClient{}
			if tt.expectErr && tt.expectErrText == "unsupported query type for streaming configuration" {
				// no call to client expected
			} else {
				mockClient.On("GetStreamingQueryConfiguration", ctx, mock.Anything).Return(tt.mockResp, tt.mockErr)
			}

			resp, err := GetStreamingQueryConfiguration(ctx, mockClient, tt.query)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrText)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResp, resp)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
