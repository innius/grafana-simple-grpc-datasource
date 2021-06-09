package models

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type DimensionKeysQuery struct {
	Filter string `json:"filter"`
}

func UnmarshalToDimensionKeysQuery(dq *backend.DataQuery) (*DimensionKeysQuery, error) {
	query := &DimensionKeysQuery{}
	if err := json.Unmarshal(dq.JSON, query); err != nil {
		return nil, err
	}

	return query, nil
}
