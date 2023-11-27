package server

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// mockCallResourceResponseSender implements backend.CallResourceResponseSender
// for use in tests.
type mockCallResourceResponseSender struct {
	response *backend.CallResourceResponse
}

// Send sets the received *backend.CallResourceResponse to s.response
func (s *mockCallResourceResponseSender) Send(response *backend.CallResourceResponse) error {
	s.response = response
	return nil
}

type backendAPIStub struct {
}

func (stub *backendAPIStub) HandleGetMetricValueQuery(ctx context.Context, query *models.MetricValueQuery) (data.Frames, error) {
	panic("not implemented") // TODO: Implement
}
func (stub *backendAPIStub) HandleGetMetricHistoryQuery(ctx context.Context, query *models.MetricHistoryQuery) (data.Frames, error) {
	panic("not implemented") // TODO: Implement
}
func (stub *backendAPIStub) HandleGetMetricAggregateQuery(ctx context.Context, query *models.MetricAggregateQuery) (data.Frames, error) {
	panic("not implemented") // TODO: Implement
}
func (stub *backendAPIStub) GetDimensionKeys(ctx context.Context, query models.GetDimensionKeysRequest) (*models.GetDimensionKeysResponse, error) {
	if query.Filter != "filter" {
		return nil, errors.New("invalid filter")
	}
	if len(query.SelectedDimensions) == 0 {
		return nil, nil
	}

	return &models.GetDimensionKeysResponse{
		Keys: []models.DimensionKeyDefinition{
			{Value: "foo", Label: "bar", Description: "bar"},
		},
	}, nil
}
func (stub *backendAPIStub) GetDimensionValues(ctx context.Context, query models.GetDimensionValuesRequest) (*models.GetDimensionValueResponse, error) {
	if query.Filter != "filter" {
		return nil, errors.New("invalid filter")
	}
	if len(query.SelectedDimensions) == 0 {
		return nil, nil
	}

	return &models.GetDimensionValueResponse{
		Values: []models.DimensionValueDefinition{
			{Value: "foo", Label: "bar", Description: "bar"},
		},
	}, nil
}
func (stub *backendAPIStub) GetMetrics(ctx context.Context, query models.GetMetricsRequest) (*models.GetMetricsResponse, error) {
	if query.Filter != "filter" {
		return nil, errors.New("invalid filter")
	}
	if len(query.Dimensions) == 0 {
		return nil, nil
	}

	return &models.GetMetricsResponse{
		Metrics: []models.MetricDefinition{
			{Value: "foo", Label: "bar", Description: "bar"},
		},
	}, nil
}

func (stub *backendAPIStub) GetQueryOptions(ctx context.Context, input models.GetQueryOptionsRequest) (*models.GetQueryOptionsResponse, error) {
	return &models.GetQueryOptionsResponse{
		Options: []models.Option{
			{
				ID:          "foo",
				Label:       "test option 1",
				Description: "test option one",
				Type:        "EnumValue",
				EnumValues: []models.EnumValue{
					{
						ID:          "foo",
						Label:       "test option 1",
						Description: "the foo option",
						Default:     true,
					},
				},
			},
		},
	}, nil
}

func (stub *backendAPIStub) Dispose() {
	panic("not implemented") // TODO: Implement
}

// TestCallResource tests CallResource calls, using backend.CallResourceRequest and backend.CallResourceResponse.
// This ensures the httpadapter for CallResource works correctly.
func TestCallResource(t *testing.T) {
	// Initialize app
	m := &backendAPIStub{}
	inst, err := newServerInstance(m)
	if err != nil {
		t.Fatalf("new app: %s", err)
	}
	if inst == nil {
		t.Fatal("inst must not be nil")
	}
	app, ok := inst.(*Server)
	if !ok {
		t.Fatal("inst must be of type *App")
	}

	// Set up and run test cases
	for _, tc := range []struct {
		name string

		method string
		path   string
		body   []byte

		expStatus int
		expBody   []byte
	}{
		{
			name:      "list query options options",
			method:    http.MethodPost,
			path:      "options",
			body:      []byte(`{"selected_options":{"foo" : "bar"}}`),
			expStatus: http.StatusOK,
			expBody:   []byte(`[{"id":"foo","label":"test option 1","description":"test option one","type":"EnumValue","enumValues":[{"id":"foo","label":"test option 1","description":"the foo option","default":true}]}]`),
		},
		{
			name:      "list query options options with empty body",
			method:    http.MethodPost,
			path:      "options",
			expStatus: http.StatusBadRequest,
		},
		{
			name:      "list query options with an invalid body",
			method:    http.MethodPost,
			path:      "options",
			expStatus: http.StatusBadRequest,
			body:      []byte(``),
		},
		{
			name:      "list dimensions",
			method:    http.MethodPost,
			path:      "dimensions",
			body:      []byte(`{"filter": "filter", "selected_dimensions": [{ "key": "foo" , "value" : "bar"}]}`),
			expStatus: http.StatusOK,
			expBody:   []byte(`[{"value":"foo","label":"bar","description":"bar"}]`),
		},
		{
			name:      "list dimensions with invalid payload",
			method:    http.MethodPost,
			path:      "dimensions",
			body:      []byte(`{"filter": "filter", "selected_dimensions": "invalid json string"}`),
			expStatus: http.StatusBadRequest,
		},
		{
			name:      "list dimensions with a backend error",
			method:    http.MethodPost,
			path:      "dimensions",
			body:      []byte(`{"filter": "with a backend error", "selected_dimensions": [{ "key": "foo" , "value" : "bar"}]}`),
			expStatus: http.StatusInternalServerError,
		},
		{
			name:      "list dimensions no backend dimensions found",
			method:    http.MethodPost,
			path:      "dimensions",
			body:      []byte(`{"filter": "filter", "selected_dimensions": []}`),
			expBody:   []byte(`[]`),
			expStatus: http.StatusOK,
		},
		{
			name:      "list dimension values",
			method:    http.MethodPost,
			path:      "dimensions/values",
			body:      []byte(`{"filter": "filter", "selected_dimensions": [{ "key": "foo" , "value" : "bar"}]}`),
			expStatus: http.StatusOK,
			expBody:   []byte(`[{"value":"foo","label":"bar","description":"bar"}]`),
		},
		{
			name:      "list dimension values with invalid payload",
			method:    http.MethodPost,
			path:      "dimensions/values",
			body:      []byte(`{"filter": "filter", "selected_dimensions": "invalid json string"}`),
			expStatus: http.StatusBadRequest,
		},
		{
			name:      "list dimensions with a backend error",
			method:    http.MethodPost,
			path:      "dimensions/values",
			body:      []byte(`{"filter": "with a backend error", "selected_dimensions": [{ "key": "foo" , "value" : "bar"}]}`),
			expStatus: http.StatusInternalServerError,
		},
		{
			name:      "list dimensions no backend dimensions found",
			method:    http.MethodPost,
			path:      "dimensions/values",
			body:      []byte(`{"filter": "filter", "selected_dimensions": []}`),
			expBody:   []byte(`[]`),
			expStatus: http.StatusOK,
		},
		{
			name:      "list metrics",
			method:    http.MethodPost,
			path:      "metrics",
			body:      []byte(`{"filter": "filter", "dimensions": [{ "key": "foo" , "value" : "bar"}]}`),
			expStatus: http.StatusOK,
			expBody:   []byte(`[{"value":"foo","label":"bar","description":"bar"}]`),
		},
		{
			name:      "list metrics with invalid payload",
			method:    http.MethodPost,
			path:      "metrics",
			body:      []byte(`{"filter": "filter", "dimensions": "invalid json string"}`),
			expStatus: http.StatusBadRequest,
		},
		{
			name:      "list metrics with a backend error",
			method:    http.MethodPost,
			path:      "metrics",
			body:      []byte(`{"filter": "with a backend error", "dimensions": [{ "key": "foo" , "value" : "bar"}]}`),
			expStatus: http.StatusInternalServerError,
		},
		{
			name:      "list metrics no backend metrics found",
			method:    http.MethodPost,
			path:      "metrics",
			body:      []byte(`{"filter": "filter", "dimensions": []}`),
			expBody:   []byte(`[]`),
			expStatus: http.StatusOK,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Request by calling CallResource. This tests the httpadapter.
			var r mockCallResourceResponseSender
			err = app.CallResource(context.Background(), &backend.CallResourceRequest{
				Method: tc.method,
				Path:   tc.path,
				Body:   tc.body,
			}, &r)
			if err != nil {
				t.Fatalf("CallResource error: %s", err)
			}
			if r.response == nil {
				t.Fatal("no response received from CallResource")
			}
			if tc.expStatus > 0 && tc.expStatus != r.response.Status {
				t.Errorf("response status should be %d, got %d", tc.expStatus, r.response.Status)
			}
			if len(tc.expBody) > 0 {
				if tb := bytes.TrimSpace(r.response.Body); !bytes.Equal(tb, tc.expBody) {
					t.Errorf("response body should be %s, got %s", tc.expBody, tb)
				}
			}
		})
	}
}
