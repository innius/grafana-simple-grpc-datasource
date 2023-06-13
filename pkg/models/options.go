package models

type Options = []Option

type EnumValue struct {
	ID          string `json:"id,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

type Option struct {
	ID          string      `json:"id,omitempty"`
	Label       string      `json:"label,omitempty"`
	Description string      `json:"description,omitempty"`
	Type        string      `json:"type,omitempty"`
	EnumValues  []EnumValue `json:"enumValues,omitempty"`
	Required    bool        `json:"required,omitempty"`
}
