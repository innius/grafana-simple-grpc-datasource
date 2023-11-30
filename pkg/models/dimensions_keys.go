package models

type GetDimensionKeysRequest struct {
	Filter             string      `json:"filter"`
	SelectedDimensions []Dimension `json:"selected_dimensions"`
}

type DimensionKeyDefinition struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type GetDimensionKeysResponse struct {
	Keys []DimensionKeyDefinition `json:"keys"`
}
