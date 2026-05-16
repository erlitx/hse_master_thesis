package dto

// JSON manifest DBT
type ManifestJSON struct {
	Nodes map[string]NodeJSON `json:"nodes"`
}

// Узел manifest
type NodeJSON struct {
	Name         string                `json:"name"`
	RelationName string                `json:"relation_name"`
	ResourceType string                `json:"resource_type"`
	Description  string                `json:"description"`
	Meta         map[string]any        `json:"meta"`
	Columns      map[string]ColumnJSON `json:"columns"`
	Refs         []RefJSON             `json:"refs"`
}

// Колонка в JSON manifest
type ColumnJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

// Ссылка в JSON manifest
type RefJSON struct {
	Name    string  `json:"name"`
	Package *string `json:"package"`
	Version *string `json:"version"`
}
