package dto

// MCPResource represents a resource fetched from the MCP server
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MIMEType    string `json:"mime_type"`
	Content     string `json:"content"` // Actual resource data (usually JSON)
}
