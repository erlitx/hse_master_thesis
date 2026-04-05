package dbt

// ManifestParser handles parsing of DBT manifest.json files
type DbtParser struct {
	manifestPath string
}

// New creates a new ManifestParser with the specified manifest file path
func New(manifestPath string) *DbtParser {
	return &DbtParser{
		manifestPath: manifestPath,
	}
}
