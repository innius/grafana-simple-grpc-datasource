package client

import (
	"bytes"
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSettingsValidations(t *testing.T) {
	t.Run("load function should ", func(t *testing.T) {
		t.Run("return an error upon empty endpoint", func(t *testing.T) {
			b := bytes.Buffer{}
			jsonData := struct{
				Endpoint string `json:"endpoint"`
			}{
				Endpoint: "",
			}
			require.NoError(t, json.NewEncoder(&b).Encode(&jsonData))
			ds := backend.DataSourceInstanceSettings{
				ID:                      1,
				JSONData:                json.RawMessage(b.Bytes()),
				DecryptedSecureJSONData: map[string]string{
					"apiKey": "",
				},
			}
			s := &BackendAPIDatasourceSettings{}
			assert.Error(t, s.Load(ds))
		})
		t.Run("return no error when endpoint is provided", func(t *testing.T) {
			b := bytes.Buffer{}
			jsonData := struct{
				Endpoint string `json:"endpoint"`
			}{
				Endpoint: "localhost:3000",
			}
			require.NoError(t, json.NewEncoder(&b).Encode(&jsonData))
			ds := backend.DataSourceInstanceSettings{
				ID:                      1,
				JSONData:                json.RawMessage(b.Bytes()),
				DecryptedSecureJSONData: map[string]string{
					"apiKey": "",
				},
			}
			s := &BackendAPIDatasourceSettings{}
			assert.NoError(t, s.Load(ds))
		})
	})

	t.Run("load function should ", func(t *testing.T) {
		b := bytes.Buffer{}
		jsonData := struct{
			Endpoint string `json:"endpoint"`
			ApiKeyAuthenticationEnabled bool `json:"apikey_authentication_enabled"`
		}{
			Endpoint: "localhost:3000",
			ApiKeyAuthenticationEnabled: true,
		}
		require.NoError(t, json.NewEncoder(&b).Encode(&jsonData))
		ds := backend.DataSourceInstanceSettings{
			ID:                      1,
			JSONData:                json.RawMessage(b.Bytes()),
			DecryptedSecureJSONData: map[string]string{
				"apiKey": "",
			},
		}
		s := &BackendAPIDatasourceSettings{}

		t.Run("return an error upon empty API Key field when api key authentication is enabled", func(t *testing.T) {
			assert.Error(t, s.Load(ds))
		})

		t.Run("return no error after API Key is provided", func(t *testing.T) {
			ds.DecryptedSecureJSONData["apiKey"] = "myApiKey"
			assert.NoError(t, s.Load(ds))
		})
	})

}
