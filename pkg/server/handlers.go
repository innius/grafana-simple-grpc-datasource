package server

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func processQueries(ctx context.Context, req *backend.QueryDataRequest, handler QueryHandlerFunc) *backend.QueryDataResponse {
	res := backend.Responses{}
	if req == nil || req.Queries == nil {
		return &backend.QueryDataResponse{
			Responses: res,
		}
	}
	for _, v := range req.Queries {
		res[v.RefID] = handler(ctx, *req, v)
	}

	return &backend.QueryDataResponse{
		Responses: res,
	}
}

func (s *Server) HandleGetMetricValueQuery(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return processQueries(ctx, req, s.handleGetMetricValueQuery), nil
}

func (s *Server) handleGetMetricValueQuery(ctx context.Context, req backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToMetricValueQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.backendAPI.HandleGetMetricValueQuery(ctx, query)
	if err != nil {
		return DataResponseErrorRequestFailed(err)
	}

	return backend.DataResponse{
		Frames: frames,
		Error:  nil,
	}
}

func (s *Server) HandleGetMetricHistoryQuery(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return processQueries(ctx, req, s.handleGetMetricHistoryQuery), nil
}

func (s *Server) handleGetMetricHistoryQuery(ctx context.Context, req backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToMetricHistoryQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.backendAPI.HandleGetMetricHistoryQuery(ctx, query)
	if err != nil {
		return DataResponseErrorRequestFailed(err)
	}

	return backend.DataResponse{
		Frames: frames,
		Error:  nil,
	}
}

func (s *Server) HandleGetMetricAggregate(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return processQueries(ctx, req, s.handleGetMetricAggregateQuery), nil
}

func (s *Server) handleGetMetricAggregateQuery(ctx context.Context, req backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToMetricAggregateQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.backendAPI.HandleGetMetricAggregateQuery(ctx, query)
	if err != nil {
		return DataResponseErrorRequestFailed(err)
	}

	return backend.DataResponse{
		Frames: frames,
		Error:  nil,
	}
}
