package connector

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
		st := status.Convert(err)
		if st.Code() == codes.InvalidArgument {
			return &models.StreamingQueryConfigurationResponse{
				LookBackPeriodLimit: 0,
				LoopInterval:        0,
				StreamingSupported:  false,
				ErrorMessage:        st.Message(),
			}, nil
		}
		// If there's an error (e.g., method not implemented), return streaming not supported
		return &models.StreamingQueryConfigurationResponse{
			LookBackPeriodLimit: 0,
			LoopInterval:        0,
			StreamingSupported:  false,
			ErrorMessage:        err.Error(),
		}, nil
	}
	if resp == nil {
		// Return streaming not supported when V4 is not available
		return &models.StreamingQueryConfigurationResponse{
			LookBackPeriodLimit: 0,
			LoopInterval:        0,
			StreamingSupported:  false,
			ErrorMessage:        "V4 API not supported by backend",
		}, nil
	}

	// Convert response
	return &models.StreamingQueryConfigurationResponse{
		LookBackPeriodLimit: time.Duration(resp.LookBackPeriodLimit) * time.Millisecond,
		LoopInterval:        time.Duration(resp.LoopInterval) * time.Millisecond,
		StreamingSupported:  resp.StreamingSupported,
		ErrorMessage:        resp.ErrorMessage,
	}, nil
}
