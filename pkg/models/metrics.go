package models

type GetMetricsRequest struct {
	Dimensions []Dimension `json:"dimensions"`
	Filter     string      `json:"filter"`
}

type MetricDefinition struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type GetMetricsResponse struct {
	Metrics []MetricDefinition `json:"metrics"`
}
