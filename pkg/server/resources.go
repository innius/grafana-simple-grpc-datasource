package server

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func (s *Server) handleGetQueryOptions(w http.ResponseWriter, r *http.Request) {
	logger := log.DefaultLogger.With("method", "getQueryOptionDefinitions")

	if r.Body == nil {
		http.Error(w, "request does not have a body", http.StatusBadRequest)
		return
	}
	// Create a JSON decoder from the request body
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	req := models.GetQueryOptionDefinitionsRequest{}
	// Use the decoder to decode the JSON into the User struct
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	res, err := s.backendAPI.GetQueryOptionDefinitions(r.Context(), req)

	if err != nil {
		logger.Error("backend returned an error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if res == nil {
		logger.Error("got no content")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	logger.Debug("returning options", "options", res.Options)

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res.Options); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleListDimensionKeys(w http.ResponseWriter, r *http.Request) {
	logger := log.DefaultLogger.With("method", "handleListDimensionKeys")

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

	res, err := s.backendAPI.HandleListDimensionsQuery(r.Context(), req)

	if err != nil {
		logger.Error("backend returned an error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (s *Server) handleListDimensionValues(w http.ResponseWriter, r *http.Request) {
	logger := log.DefaultLogger.With("method", "handleListDimensionValues")

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

	res, err := s.backendAPI.HandleListDimensionValuesQuery(r.Context(), req)
	if err != nil {
		logger.Error("backend returned an error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (s *Server) handleListMetrics(w http.ResponseWriter, r *http.Request) {
	logger := log.DefaultLogger.With("method", "handleListMetrics")

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

	res, err := s.backendAPI.HandleListMetricsQuery(r.Context(), req)
	if err != nil {
		logger.Error("backend returned an error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (a *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/options", a.handleGetQueryOptions)
	mux.HandleFunc("/dimensions", a.handleListDimensionKeys)
	mux.HandleFunc("/dimensions/values", a.handleListDimensionValues)
	mux.HandleFunc("/metrics", a.handleListMetrics)
}
