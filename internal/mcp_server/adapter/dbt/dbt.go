package dbt

// Парсер manifest DBT
type DbtParser struct {
	manifestPath string
}

// Создаёт новый экземпляр
func New(manifestPath string) *DbtParser {
	return &DbtParser{
		manifestPath: manifestPath,
	}
}
