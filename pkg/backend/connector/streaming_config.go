package connector

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	v4 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v4"
)

// GetStreamingQueryConfiguration retrieves streaming configuration from the backend
func GetStreamingQueryConfiguration(ctx context.Context, client client.BackendAPIClient, query models.StreamingQueryConfigurationRequest) (*models.StreamingQueryConfigurationResponse, error) {
	req := &v4.GetStreamingQueryConfigurationRequest{}
	// Convert the query to the appropriate v4 request type
	switch q := query.Query.(type) {
	case models.MetricValueQuery:
		req.Query = &v4.GetStreamingQueryConfigurationRequest_GetMetricValueRequest{
			GetMetricValueRequest: convertToMetricValueRequest(q),
		}
	case models.MetricHistoryQuery:
		req.Query = &v4.GetStreamingQueryConfigurationRequest_GetMetricHistoryRequest{
			GetMetricHistoryRequest: convertToMetricHistoryRequest(q),
		}
	case models.MetricAggregateQuery:
		req.Query = &v4.GetStreamingQueryConfigurationRequest_GetMetricAggregateRequest{
			GetMetricAggregateRequest: convertToMetricAggregateRequest(q),
		}
	default:
		return nil, errors.Errorf("unsupported query type for streaming configuration: %T", q)
	}

	resp, err := client.GetStreamingQueryConfiguration(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get streaming query configuration")
	}
	if resp == nil {
		// Fallback to default configuration if v4 is not supported
		return &models.StreamingQueryConfigurationResponse{
			LookBackPeriodLimit: time.Hour,
			LoopInterval:        10 * time.Second,
		}, nil
	}

	// Convert response
	return &models.StreamingQueryConfigurationResponse{
		LookBackPeriodLimit: time.Duration(resp.LookBackPeriodLimit) * time.Millisecond,
		LoopInterval:        time.Duration(resp.LoopInterval) * time.Millisecond,
	}, nil
}
