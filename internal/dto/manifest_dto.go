package dto

// manifestJSON represents the structure of the manifest.json file
type ManifestJSON struct {
	Nodes map[string]NodeJSON `json:"nodes"`
}

// nodeJSON represents a node in the manifest
type NodeJSON struct {
	Name         string                `json:"name"`
	RelationName string                `json:"relation_name"`
	ResourceType string                `json:"resource_type"`
	Description  string                `json:"description"`
	Meta         map[string]any        `json:"meta"`
	Columns      map[string]ColumnJSON `json:"columns"`
	Refs         []RefJSON             `json:"refs"`
}

// columnJSON represents a column definition
type ColumnJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

// refJSON represents a model reference
type RefJSON struct {
	Name    string  `json:"name"`
	Package *string `json:"package"`
	Version *string `json:"version"`
}
