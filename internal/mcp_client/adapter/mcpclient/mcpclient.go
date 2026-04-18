package mcpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type Client struct {
	httpClient *http.Client
	serverURL  string
	requestID  atomic.Int64
}

func New(serverURL string) (*Client, error) {
	return &Client{
		httpClient: &http.Client{},
		serverURL:  serverURL,
	}, nil
}

// ListResources fetches all available resources from the MCP server
func (c *Client) ListResources(ctx context.Context) ([]dto.MCPResource, error) {
	log.Debug().Str("server_url", c.serverURL).Msg("listing resources from MCP server")

	// Call resources/list using MCP types
	var listResult mcp.ListResourcesResult
	if err := c.call(ctx, "resources/list", nil, &listResult); err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	log.Debug().Int("count", len(listResult.Resources)).Msg("fetched resource list")

	// Convert MCP resources to our DTO format and fetch content
	resources := make([]dto.MCPResource, 0, len(listResult.Resources))
	for _, res := range listResult.Resources {
		// Read the actual resource content
		content, err := c.ReadResource(ctx, res.URI)
		if err != nil {
			log.Warn().Err(err).Str("uri", res.URI).Msg("failed to read resource, skipping")
			continue
		}
		resources = append(resources, *content)
	}

	log.Info().Int("count", len(resources)).Msg("successfully fetched all resources")
	return resources, nil
}

// ReadResource fetches a specific resource by URI
func (c *Client) ReadResource(ctx context.Context, uri string) (*dto.MCPResource, error) {
	log.Debug().Str("uri", uri).Msg("reading resource from MCP server")

	// Call resources/read using custom unmarshaling
	params := struct {
		URI string `json:"uri"`
	}{
		URI: uri,
	}

	// Use a custom result structure to handle the contents manually
	var rawResult struct {
		Contents []map[string]any `json:"contents"`
	}
	if err := c.call(ctx, "resources/read", params, &rawResult); err != nil {
		return nil, fmt.Errorf("failed to read resource %s: %w", uri, err)
	}

	if len(rawResult.Contents) == 0 {
		return nil, fmt.Errorf("no content returned for resource %s", uri)
	}

	// Parse the first content using mcp.ParseResourceContents
	firstContent, err := mcp.ParseResourceContents(rawResult.Contents[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource contents: %w", err)
	}

	// Extract text content based on the type
	var contentText string
	var mimeType string

	// Try to use AsTextResourceContents helper
	if textContent, ok := mcp.AsTextResourceContents(firstContent); ok {
		contentText = textContent.Text
		mimeType = textContent.MIMEType
	} else {
		// Debug: log the actual type
		log.Warn().Msgf("unexpected content type: %T", firstContent)
		return nil, fmt.Errorf("unexpected content type for resource %s (got %T)", uri, firstContent)
	}

	resource := &dto.MCPResource{
		URI:      uri,
		MIMEType: mimeType,
		Content:  contentText,
	}

	log.Debug().Str("uri", uri).Int("content_length", len(contentText)).Msg("successfully read resource")
	return resource, nil
}

// ListTools fetches all available tools from the MCP server.
func (c *Client) ListTools(ctx context.Context) ([]dto.MCPToolInfo, error) {
	log.Debug().Str("server_url", c.serverURL).Msg("listing tools from MCP server")

	var rawResult struct {
		Tools []dto.MCPToolInfo `json:"tools"`
	}
	if err := c.call(ctx, "tools/list", nil, &rawResult); err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	log.Info().Int("count", len(rawResult.Tools)).Msg("successfully fetched tools")
	return rawResult.Tools, nil
}

// CallTool executes a specific MCP tool with JSON-RPC tools/call.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (map[string]interface{}, error) {
	log.Debug().Str("tool_name", name).Msg("calling MCP tool")

	params := dto.MCPToolCallRequest{
		Name:      name,
		Arguments: arguments,
	}

	var result map[string]interface{}
	if err := c.call(ctx, "tools/call", params, &result); err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", name, err)
	}

	return result, nil
}

func (c *Client) call(ctx context.Context, method string, params interface{}, result interface{}) error {
	requestID := c.requestID.Add(1)

	type jsonRPCRequest struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      int64       `json:"id"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params,omitempty"`
	}

	rpcRequest := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      requestID,
		Method:  method,
		Params:  params,
	}

	requestBody, err := json.Marshal(rpcRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Debug().Msgf("----------BODY: %v", string(requestBody))
	log.Debug().
		Str("method", method).
		Int64("request_id", requestID).
		Str("url", c.serverURL).
		Msg("sending MCP request")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.serverURL, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: status=%d, body=%s", resp.StatusCode, string(responseBody))
	}

	var rpcError mcp.JSONRPCError
	if err := json.Unmarshal(responseBody, &rpcError); err == nil && rpcError.Error.Code != 0 {
		return fmt.Errorf("JSON-RPC error: code=%d, message=%s", rpcError.Error.Code, rpcError.Error.Message)
	}

	var rpcResponse mcp.JSONRPCResponse
	if err := json.Unmarshal(responseBody, &rpcResponse); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result != nil && rpcResponse.Result != nil {
		resultBytes, err := json.Marshal(rpcResponse.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		if err := json.Unmarshal(resultBytes, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	log.Debug().
		Str("method", method).
		Int64("request_id", requestID).
		Msg("received MCP response")

	return nil
}

// Close closes the MCP client connection
func (c *Client) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}
