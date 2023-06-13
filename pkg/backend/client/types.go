package client

import (
	"encoding/json"
	"fmt"

	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type BackendAPIClient interface {
	v3.GrafanaQueryAPIClient
	Dispose()
}

type BackendAPIDatasourceSettings struct {
	ID         string `json:"-"`
	Endpoint   string `json:"endpoint"`
	APIKey     string `json:"-"`
	MaxRetries uint   `json:"max_retries"`
}

func (s *BackendAPIDatasourceSettings) Load(config backend.DataSourceInstanceSettings) error {
	if config.JSONData != nil && len(config.JSONData) > 1 {
		if err := json.Unmarshal(config.JSONData, s); err != nil {
			return fmt.Errorf("could not unmarshal DatasourceSettings json: %w", err)
		}
	}
	s.ID = config.UID
	s.APIKey = config.DecryptedSecureJSONData["apiKey"]
	return nil
}
