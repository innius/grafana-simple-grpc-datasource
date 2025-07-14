package plugin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
)

func TestSubscribeStreamIntegration(t *testing.T) {
	// Create mock backend API
	mockAPI := &MockBackendAPI{}

	// Create test frame
	testFrame := data.NewFrame("test-frame",
		data.NewField("time", nil, []time.Time{time.Now().Add(-1 * time.Hour), time.Now()}),
		data.NewField("value", nil, []float64{1.0, 2.0}),
	)
	// Create test query
	testQuery := Q{
		QueryType: models.QueryMetricHistory,
		MetricBaseQuery: models.MetricBaseQuery{
			Metrics: []models.Metric{{MetricId: "test-metric"}},
		},
	}

	// Setup mock expectations
	mockAPI.On("HandleGetMetricHistoryQuery", mock.Anything, mock.Anything).Return(data.Frames{testFrame}, nil)
	mockAPI.On("GetStreamingQueryConfiguration", mock.Anything, mock.Anything).
		Return(&models.StreamingQueryConfigurationResponse{
			LookBackPeriodLimit: time.Hour,
			LoopInterval:        time.Minute,
		}, nil)
	// Create datasource with mock API
	datasource := &Datasource{
		backendAPI: mockAPI,
	}

	queryData, err := json.Marshal(testQuery)
	assert.NoError(t, err)

	// Test SubscribeStream
	ctx := context.Background()
	req := &backend.SubscribeStreamRequest{
		Path: "test-path",
		Data: queryData,
	}

	resp, err := datasource.SubscribeStream(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, backend.SubscribeStreamStatusOK, resp.Status)
	assert.NotNil(t, resp.InitialData)

	// Verify mock was called
	mockAPI.AssertExpectations(t)
}

func TestSubscribeStreamValidationFailure(t *testing.T) {
	// Create mock backend API
	mockAPI := &MockBackendAPI{}

	// Create datasource with mock API
	datasource := &Datasource{
		backendAPI: mockAPI,
	}

	// Create invalid query (missing query type)
	testQuery := Q{
		MetricBaseQuery: models.MetricBaseQuery{
			Metrics: []models.Metric{{MetricId: "test-metric"}},
		},
	}

	queryData, err := json.Marshal(testQuery)
	assert.NoError(t, err)

	// Test SubscribeStream with invalid query
	ctx := context.Background()
	req := &backend.SubscribeStreamRequest{
		Path: "test-path",
		Data: queryData,
	}

	resp, err := datasource.SubscribeStream(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, backend.SubscribeStreamStatusNotFound, resp.Status)
	assert.Nil(t, resp.InitialData)

	// Verify no backend calls were made
	mockAPI.AssertExpectations(t)
}

func TestRunStreamQueryValidation(t *testing.T) {
	// Create mock backend API
	mockAPI := &MockBackendAPI{}

	// Create datasource with mock API
	datasource := &Datasource{
		backendAPI: mockAPI,
	}

	// Test with invalid query data
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := &backend.RunStreamRequest{
		Data: []byte("invalid json"),
	}

	// Create a mock packet sender for StreamSender
	mockPacketSender := &MockStreamPacketSender{}
	sender := backend.NewStreamSender(mockPacketSender)

	err := datasource.RunStream(ctx, req, sender)

	// Should return parsing error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")

	// Verify no backend calls were made
	mockAPI.AssertExpectations(t)
}

// MockStreamPacketSender for testing StreamSender
type MockStreamPacketSender struct {
	mock.Mock
}

func (m *MockStreamPacketSender) Send(packet *backend.StreamPacket) error {
	args := m.Called(packet)
	return args.Error(0)
}
