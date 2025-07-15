package models

import "time"

// StreamingQueryConfigurationRequest represents a request to get streaming configuration from backend
type StreamingQueryConfigurationRequest struct {
	// The query for which streaming configuration is requested
	Query interface{} `json:"query"`
}

// StreamingQueryConfigurationResponse represents the streaming configuration returned by backend
type StreamingQueryConfigurationResponse struct {
	// LookBackPeriodLimit in milliseconds - maximum allowed lookback period for initial dataset
	LookBackPeriodLimit time.Duration `json:"lookBackPeriod"`
	// LoopInterval in milliseconds - polling interval for RunStream
	LoopInterval time.Duration `json:"loopInterval"`
	// StreamingSupported indicates whether streaming is supported for this query
	StreamingSupported bool `json:"streamingSupported"`
	// ErrorMessage provides a reason if streaming is not supported
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// // StreamingConfiguration holds the resolved streaming configuration
// type StreamingConfiguration struct {
// 	// LookBackPeriod as duration - resolved from backend and client preferences
// 	LookBackPeriod time.Duration
// 	// LoopInterval as duration - polling interval for streaming
// 	LoopInterval time.Duration
// 	// MaxLookBackPeriod as duration - maximum allowed by backend
// 	MaxLookBackPeriod time.Duration
// }
//
// // ResolveStreamingConfig resolves the final streaming configuration
// // taking into account backend limits and client preferences
// func ResolveStreamingConfig(backendConfig *StreamingQueryConfigurationResponse, clientLookBackPeriod *string) StreamingConfiguration {
// 	config := StreamingConfiguration{
// 		LoopInterval:      backendConfig.LoopInterval,
// 		MaxLookBackPeriod: backendConfig.LookBackPeriodLimit,
// 	}
//
// 	// Default lookback period
// 	defaultLookBack := 1 * time.Hour
// 	if config.MaxLookBackPeriod > 0 && config.MaxLookBackPeriod < defaultLookBack {
// 		defaultLookBack = config.MaxLookBackPeriod
// 	}
//
// 	// Resolve client preference
// 	if clientLookBackPeriod != nil && *clientLookBackPeriod != "" {
// 		if clientDuration, err := time.ParseDuration(*clientLookBackPeriod); err == nil {
// 			// Use client preference but respect backend maximum
// 			if config.MaxLookBackPeriod > 0 && clientDuration > config.MaxLookBackPeriod {
// 				config.LookBackPeriod = config.MaxLookBackPeriod
// 			} else {
// 				config.LookBackPeriod = clientDuration
// 			}
// 		} else {
// 			config.LookBackPeriod = defaultLookBack
// 		}
// 	} else {
// 		config.LookBackPeriod = defaultLookBack
// 	}
//
// 	return config
// }
