package models

type Options = []Option

type EnumValue struct {
	ID          string `json:"id,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	Default     bool   `json:"default,omitempty"`
}

type Option struct {
	ID          string      `json:"id,omitempty"`
	Label       string      `json:"label,omitempty"`
	Description string      `json:"description,omitempty"`
	Type        string      `json:"type,omitempty"`
	EnumValues  []EnumValue `json:"enumValues,omitempty"`
	Required    bool        `json:"required,omitempty"`
}

// GetQueryOptionsRequest defines the request for the GetQueryOptions endpoint.
type GetQueryOptionsRequest struct {
	// QueryType is the selected query type
	QueryType string `json:"query_type,omitempty"`

	// SelectedOptions are the options which are currently selected for the query
	SelectedOptions map[string]string `json:"selected_options,omitempty"`
}

// GetQueryOptionsResponse defines the response for the GetQueryOptions[Definitions] endpoint
type GetQueryOptionsResponse struct {
	Options Options `json:"options,omitempty"`
}
