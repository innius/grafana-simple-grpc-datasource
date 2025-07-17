package models

import "time"

// StreamingQueryConfigurationRequest represents a request to get streaming configuration from backend
type StreamingQueryConfigurationRequest struct {
	// The query for which streaming configuration is requested
	Query any `json:"query"`
}

// StreamingQueryConfigurationResponse represents the streaming configuration returned by backend
type StreamingQueryConfigurationResponse struct {
	// LookBackPeriod in milliseconds - server-recommended lookback period for initial dataset
	// This serves as both the default value when client doesn't specify one, and the maximum
	// allowed value that client requests cannot exceed
	LookBackPeriod time.Duration `json:"lookBackPeriod"`
	// LoopInterval in milliseconds - polling interval for RunStream
	LoopInterval time.Duration `json:"loopInterval"`
	// StreamingSupported indicates whether streaming is supported for this query
	StreamingSupported bool `json:"streamingSupported"`
	// ErrorMessage provides a reason if streaming is not supported
	ErrorMessage string `json:"errorMessage,omitempty"`
}
