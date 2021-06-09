package models

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type DimensionValueQuery struct {
	DimensionKey string `json:"dimension_key"`
	Filter string `json:"filter"`
}

func UnmarshalToDimensionValueQuery(dq *backend.DataQuery) (*DimensionValueQuery, error) {
	query := &DimensionValueQuery{}
	if err := json.Unmarshal(dq.JSON, query); err != nil {
		return nil, err
	}

	return query, nil
}
