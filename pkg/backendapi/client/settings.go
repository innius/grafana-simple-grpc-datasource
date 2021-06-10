package client

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type BackendAPIDatasourceSettings struct {
	ID                          string `json:"-"`
	Endpoint                    string `json:"endpoint"`
	APIKey                      string `json:"-"`
	ApiKeyAuthenticationEnabled bool   `json:"apikey_authentication_enabled"`
}

func (s *BackendAPIDatasourceSettings) Load(config backend.DataSourceInstanceSettings) error {
	if config.JSONData != nil && len(config.JSONData) > 1 {
		if err := json.Unmarshal(config.JSONData, s); err != nil {
			return fmt.Errorf("could not unmarshal DatasourceSettings json: %w", err)
		}
	}
	s.ID = config.UID
	s.APIKey = config.DecryptedSecureJSONData["apiKey"]

	return validate(s)
}

func validate(s *BackendAPIDatasourceSettings) error {
	if s.Endpoint == "" {
		return fmt.Errorf("endpoint is a required configuration setting")
	}
	if s.ApiKeyAuthenticationEnabled && s.APIKey == "" {
		return fmt.Errorf("API Key is required when API Key authentication is enabled")
	}
	return nil
}
