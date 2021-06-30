package server

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func processQueries(ctx context.Context, req *backend.QueryDataRequest, handler QueryHandlerFunc) *backend.QueryDataResponse {
	res := backend.Responses{}
	for _, v := range req.Queries {
		res[v.RefID] = handler(ctx, req, v)
	}

	return &backend.QueryDataResponse{
		Responses: res,
	}

}

func (s *Server) HandleGetMetricValueQuery(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return processQueries(ctx, req, s.handleGetMetricValueQuery), nil
}

func (s *Server) handleGetMetricValueQuery(ctx context.Context, req *backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToMetricValueQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.Datasource.HandleGetMetricValueQuery(ctx, query)
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

func (s *Server) handleGetMetricHistoryQuery(ctx context.Context, req *backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToMetricHistoryQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.Datasource.HandleGetMetricHistoryQuery(ctx, query)
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

func (s *Server) handleGetMetricAggregateQuery(ctx context.Context, req *backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToMetricAggregateQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.Datasource.HandleGetMetricAggregateQuery(ctx, query)
	if err != nil {
		return DataResponseErrorRequestFailed(err)
	}

	return backend.DataResponse{
		Frames: frames,
		Error:  nil,
	}
}

func (s *Server) HandleListDimensionKeysQuery(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return processQueries(ctx, req, s.handleListDimensionKeysQuery), nil
}

func (s *Server) handleListDimensionKeysQuery(ctx context.Context, req *backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToDimensionKeysQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.Datasource.HandleListDimensionsQuery(ctx, query)
	if err != nil {
		return DataResponseErrorRequestFailed(err)
	}

	return backend.DataResponse{
		Frames: frames,
		Error:  nil,
	}
}

func (s *Server) HandleListDimensionValuesQuery(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return processQueries(ctx, req, s.handleListDimensionValuesQuery), nil
}

func (s *Server) handleListDimensionValuesQuery(ctx context.Context, req *backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToDimensionValueQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.Datasource.HandleListDimensionValuesQuery(ctx, query)
	if err != nil {
		return DataResponseErrorRequestFailed(err)
	}

	return backend.DataResponse{
		Frames: frames,
		Error:  nil,
	}
}

func (s *Server) HandleListMetricsQuery(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return processQueries(ctx, req, s.handleListMetricsQuery), nil
}

func (s *Server) handleListMetricsQuery(ctx context.Context, req *backend.QueryDataRequest, q backend.DataQuery) backend.DataResponse {
	query, err := models.UnmarshalToMetricsQuery(&q)
	if err != nil {
		return DataResponseErrorUnmarshal(err)
	}

	frames, err := s.Datasource.HandleListMetricsQuery(ctx, query)
	if err != nil {
		return DataResponseErrorRequestFailed(err)
	}

	return backend.DataResponse{
		Frames: frames,
		Error:  nil,
	}
}
