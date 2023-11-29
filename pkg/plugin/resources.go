package plugin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
)

// httpStatusFromCode translates the given GRPC code into an HTTP
// response. This is used to set the HTTP status code for unary RPCs.
// (Streaming RPCs cannot convey a GRPC status code until the stream
// completes, so they use a 200 HTTP status code and then encode the
// actual status, along with any trailer metadata, at the end of the
// response stream.)
func httpStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusBadGateway
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusUnprocessableEntity
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func renderError(ctx context.Context, st *status.Status, w http.ResponseWriter) {
	// return context cancellation error as a 499 if the grpc server returned a
	// Canceled or DeadlineExceeded error and the current context is canceled
	if (st.Code() == codes.Canceled || st.Code() == codes.DeadlineExceeded) && ctx.Err() != nil {
		http.Error(w, "Client Closed Request", 499)
		return
	}
	code := httpStatusFromCode(st.Code())

	http.Error(w, st.Message(), code)
}

func (s *Datasource) handleGetQueryOptions(w http.ResponseWriter, r *http.Request) {
	logger := log.DefaultLogger.With("method", "getQueryOptionDefinitions")

	if r.Body == nil {
		http.Error(w, "request does not have a body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req models.GetQueryOptionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	res, err := s.backendAPI.GetQueryOptions(r.Context(), req)
	if err != nil {
		logger.Error("backend returned an error", "error", err.Error())
		renderError(r.Context(), status.Convert(err), w)
		return
	}

	options := []models.Option{}
	if res != nil && res.Options != nil {
		options = res.Options
	}

	logger.Debug("returning options", "options", res.Options)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(options); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Datasource) handleGetDimensionKeys(w http.ResponseWriter, r *http.Request) {
	logger := log.DefaultLogger.With("method", "handleGetDimensionKeys")

	if r.Body == nil {
		http.Error(w, "request does not have a body", http.StatusBadRequest)
		return
	}
	// Create a JSON decoder from the request body
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	req := models.GetDimensionKeysRequest{}
	// Use the decoder to decode the JSON into the User struct
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	res, err := s.backendAPI.GetDimensionKeys(r.Context(), req)

	if err != nil {
		logger.Error("backend returned an error", "error", err.Error())
		renderError(r.Context(), status.Convert(err), w)
		return
	}

	keys := []models.DimensionKeyDefinition{}
	if res != nil && res.Keys != nil {
		keys = res.Keys
	}

	logger.Debug("returning keys", "keys", keys)

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(keys); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Datasource) handleGetDimensionValues(w http.ResponseWriter, r *http.Request) {
	logger := log.DefaultLogger.With("method", "handleGetDimensionValues")

	if r.Body == nil {
		http.Error(w, "request does not have a body", http.StatusBadRequest)
		return
	}
	// Create a JSON decoder from the request body
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	req := models.GetDimensionValuesRequest{}
	// Use the decoder to decode the JSON into the User struct
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	res, err := s.backendAPI.GetDimensionValues(r.Context(), req)

	if err != nil {
		renderError(r.Context(), status.Convert(err), w)
		logger.Error("backend returned an error", "error", err.Error())
		return
	}

	values := []models.DimensionValueDefinition{}
	if res != nil && res.Values != nil {
		values = res.Values
	}

	logger.Debug("returning values", "values", values)

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(values); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Datasource) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	logger := log.DefaultLogger.With("method", "handleGetMetrics")

	if r.Body == nil {
		http.Error(w, "request does not have a body", http.StatusBadRequest)
		return
	}
	// Create a JSON decoder from the request body
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	req := models.GetMetricsRequest{}
	// Use the decoder to decode the JSON into the User struct
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	res, err := s.backendAPI.GetMetrics(r.Context(), req)
	if err != nil {
		logger.Error("backend returned an error", "error", err.Error())
		renderError(r.Context(), status.Convert(err), w)
		return
	}

	values := []models.MetricDefinition{}
	if res != nil && res.Metrics != nil {
		values = res.Metrics
	}

	logger.Debug("returning values", "values", values)

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(values); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *Datasource) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/options", a.handleGetQueryOptions)
	mux.HandleFunc("/dimensions", a.handleGetDimensionKeys)
	mux.HandleFunc("/dimensions/values", a.handleGetDimensionValues)
	mux.HandleFunc("/metrics", a.handleGetMetrics)
}
