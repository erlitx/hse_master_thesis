package domain

// Manifest DBT
type DBTManifest struct {
	Models []DBTModel `json:"models"`
}

// Модель DBT
type DBTModel struct {
	Name         string           `json:"name"`
	RelationName string           `json:"relation_name"`
	Description  string           `json:"description"`
	Meta         map[string]any   `json:"meta"`
	Columns      []DBTModelColumn `json:"columns"`
	Refs         []DBTModelRef    `json:"refs"`
}

// Колонка модели DBT
type DBTModelColumn struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

// Ссылка модели DBT
type DBTModelRef struct {
	Name    string  `json:"name"`
	Package *string `json:"package,omitempty"`
	Version *string `json:"version,omitempty"`
}
