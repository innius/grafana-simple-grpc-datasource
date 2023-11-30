package models

type GetDimensionValuesRequest struct {
	DimensionKey       string      `json:"dimensionKey"`
	Filter             string      `json:"filter"`
	SelectedDimensions []Dimension `json:"selected_dimensions"`
}

type DimensionValueDefinition struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type GetDimensionValueResponse struct {
	Values []DimensionValueDefinition `json:"values"`
}
