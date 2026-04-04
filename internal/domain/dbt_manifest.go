package domain

// DBTManifest represents the parsed DBT manifest data
type DBTManifest struct {
	Models []DBTModel `json:"models"`
}

// DBTModel represents a single DBT model with its metadata
type DBTModel struct {
	Name         string            `json:"name"`
	RelationName string            `json:"relation_name"`
	Description  string            `json:"description"`
	Meta         map[string]any    `json:"meta"`
	Columns      []DBTModelColumn  `json:"columns"`
	Refs         []DBTModelRef     `json:"refs"`
}

// DBTModelColumn represents a column in a DBT model
type DBTModelColumn struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

// DBTModelRef represents a reference to another DBT model
type DBTModelRef struct {
	Name    string  `json:"name"`
	Package *string `json:"package,omitempty"`
	Version *string `json:"version,omitempty"`
}
