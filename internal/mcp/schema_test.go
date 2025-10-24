package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToolDefinitions(t *testing.T) {
	tools := GetToolDefinitions()

	// Verify we have all expected tools
	assert.Len(t, tools, 6, "should have 6 tool definitions")

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	assert.True(t, toolNames[ToolContextSearch], "should have context.search tool")
	assert.True(t, toolNames[ToolContextGetRelatedInfo], "should have context.get_related_info tool")
	assert.True(t, toolNames[ToolContextIndexControl], "should have context.index_control tool")
	assert.True(t, toolNames[ToolContextConnectorManagement], "should have context.connector_management tool")
	assert.True(t, toolNames[ToolContextExplain], "should have context.explain tool")
	assert.True(t, toolNames[ToolContextGrep], "should have context.grep tool")
}

func TestToolDefinition_ContextSearch(t *testing.T) {
	tools := GetToolDefinitions()

	var searchTool *ToolDefinition
	for i := range tools {
		if tools[i].Name == ToolContextSearch {
			searchTool = &tools[i]
			break
		}
	}

	require.NotNil(t, searchTool, "context.search tool should exist")
	assert.Equal(t, ToolContextSearch, searchTool.Name)
	assert.NotEmpty(t, searchTool.Description)
	assert.NotNil(t, searchTool.InputSchema)

	// Verify schema is valid JSON
	var schemaMap map[string]interface{}
	err := json.Unmarshal(searchTool.InputSchema, &schemaMap)
	require.NoError(t, err)

	// Verify required schema fields
	assert.Equal(t, "object", schemaMap["type"])

	properties, ok := schemaMap["properties"].(map[string]interface{})
	require.True(t, ok, "schema should have properties")
	assert.Contains(t, properties, "query")
	assert.Contains(t, properties, "top_k")
	assert.Contains(t, properties, "filters")
	assert.Contains(t, properties, "work_context")

	required, ok := schemaMap["required"].([]interface{})
	require.True(t, ok, "schema should have required fields")
	assert.Contains(t, required, "query")
}

func TestToolDefinition_GetRelatedInfo(t *testing.T) {
	tools := GetToolDefinitions()

	var relatedTool *ToolDefinition
	for i := range tools {
		if tools[i].Name == ToolContextGetRelatedInfo {
			relatedTool = &tools[i]
			break
		}
	}

	require.NotNil(t, relatedTool, "context.get_related_info tool should exist")
	assert.Equal(t, ToolContextGetRelatedInfo, relatedTool.Name)
	assert.NotEmpty(t, relatedTool.Description)

	// Verify schema has file_path and ticket_id
	var schemaMap map[string]interface{}
	err := json.Unmarshal(relatedTool.InputSchema, &schemaMap)
	require.NoError(t, err)

	properties, ok := schemaMap["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "file_path")
	assert.Contains(t, properties, "ticket_id")
}

func TestToolDefinition_IndexControl(t *testing.T) {
	tools := GetToolDefinitions()

	var indexTool *ToolDefinition
	for i := range tools {
		if tools[i].Name == ToolContextIndexControl {
			indexTool = &tools[i]
			break
		}
	}

	require.NotNil(t, indexTool, "context.index_control tool should exist")
	assert.Equal(t, ToolContextIndexControl, indexTool.Name)
	assert.NotEmpty(t, indexTool.Description)

	// Verify schema has action field
	var schemaMap map[string]interface{}
	err := json.Unmarshal(indexTool.InputSchema, &schemaMap)
	require.NoError(t, err)

	properties, ok := schemaMap["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "action")
}

func TestToolDefinition_ConnectorManagement(t *testing.T) {
	tools := GetToolDefinitions()

	var connectorTool *ToolDefinition
	for i := range tools {
		if tools[i].Name == ToolContextConnectorManagement {
			connectorTool = &tools[i]
			break
		}
	}

	require.NotNil(t, connectorTool, "context.connector_management tool should exist")
	assert.Equal(t, ToolContextConnectorManagement, connectorTool.Name)
	assert.NotEmpty(t, connectorTool.Description)

	// Verify schema has action field
	var schemaMap map[string]interface{}
	err := json.Unmarshal(connectorTool.InputSchema, &schemaMap)
	require.NoError(t, err)

	properties, ok := schemaMap["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "action")
}

func TestResourceDefinition_Constants(t *testing.T) {
	assert.Equal(t, "engine", ResourceScheme)
	assert.Equal(t, "files", ResourceFiles)
}

func TestSearchRequest_JSONSerialization(t *testing.T) {
	req := SearchRequest{
		Query: "test query",
		TopK:  10,
		Filters: &SearchFilters{
			SourceTypes: []string{"file", "github"},
			DateRange: &DateRange{
				From: "2024-01-01",
				To:   "2024-12-31",
			},
		},
		WorkContext: &WorkContext{
			ActiveFile:    "main.go",
			GitBranch:     "main",
			OpenTicketIDs: []string{"TASK-123"},
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded SearchRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.Query, decoded.Query)
	assert.Equal(t, req.TopK, decoded.TopK)
	assert.Equal(t, req.Filters.SourceTypes, decoded.Filters.SourceTypes)
	assert.Equal(t, req.WorkContext.ActiveFile, decoded.WorkContext.ActiveFile)
	assert.Equal(t, req.WorkContext.GitBranch, decoded.WorkContext.GitBranch)
}

func TestSearchResponse_JSONSerialization(t *testing.T) {
	resp := SearchResponse{
		Results: []SearchResultItem{
			{
				ID:         "test-1",
				Content:    "test content",
				SourceType: "file",
				Score:      0.95,
				Metadata: map[string]interface{}{
					"line": 10,
				},
			},
		},
		TotalCount: 1,
		QueryTime:  100.5,
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded SearchResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Results, 1)
	assert.Equal(t, "test content", decoded.Results[0].Content)
	assert.Equal(t, float32(0.95), decoded.Results[0].Score)
	assert.Equal(t, 100.5, decoded.QueryTime)
}

func TestGetRelatedInfoRequest_JSONSerialization(t *testing.T) {
	req := GetRelatedInfoRequest{
		FilePath: "main.go",
		TicketID: "TASK-123",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded GetRelatedInfoRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.FilePath, decoded.FilePath)
	assert.Equal(t, req.TicketID, decoded.TicketID)
}

func TestIndexControlRequest_JSONSerialization(t *testing.T) {
	req := IndexControlRequest{
		Action:     "status",
		Connectors: []string{"github", "slack"},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded IndexControlRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "status", decoded.Action)
	assert.Equal(t, []string{"github", "slack"}, decoded.Connectors)
}

func TestConnectorManagementRequest_JSONSerialization(t *testing.T) {
	req := ConnectorManagementRequest{
		Action:      "add",
		ConnectorID: "github-main",
		ConnectorConfig: map[string]interface{}{
			"token": "secret",
			"repo":  "owner/repo",
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded ConnectorManagementRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "add", decoded.Action)
	assert.Equal(t, "github-main", decoded.ConnectorID)
	assert.Equal(t, "secret", decoded.ConnectorConfig["token"])
}

func TestExplainRequest_JSONSerialization(t *testing.T) {
	req := ExplainRequest{
		Target:  "authentication system",
		Context: "how users log in",
		Depth:   "comprehensive",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded ExplainRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.Target, decoded.Target)
	assert.Equal(t, req.Context, decoded.Context)
	assert.Equal(t, req.Depth, decoded.Depth)
}

func TestGrepRequest_JSONSerialization(t *testing.T) {
	req := GrepRequest{
		Pattern:         "func.*Auth",
		Path:            "/src",
		Include:         "*.go",
		CaseInsensitive: true,
		Context:         3,
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded GrepRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.Pattern, decoded.Pattern)
	assert.Equal(t, req.Path, decoded.Path)
	assert.Equal(t, req.Include, decoded.Include)
	assert.Equal(t, req.CaseInsensitive, decoded.CaseInsensitive)
	assert.Equal(t, req.Context, decoded.Context)
}

func TestToolDefinition_ContextExplain(t *testing.T) {
	tools := GetToolDefinitions()

	var explainTool *ToolDefinition
	for i := range tools {
		if tools[i].Name == ToolContextExplain {
			explainTool = &tools[i]
			break
		}
	}

	require.NotNil(t, explainTool, "context.explain tool should exist")
	assert.Equal(t, ToolContextExplain, explainTool.Name)
	assert.NotEmpty(t, explainTool.Description)

	// Verify schema has target field
	var schemaMap map[string]interface{}
	err := json.Unmarshal(explainTool.InputSchema, &schemaMap)
	require.NoError(t, err)

	properties, ok := schemaMap["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "target")
	assert.Contains(t, properties, "depth")

	required, ok := schemaMap["required"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, required, "target")
}

func TestToolDefinition_ContextGrep(t *testing.T) {
	tools := GetToolDefinitions()

	var grepTool *ToolDefinition
	for i := range tools {
		if tools[i].Name == ToolContextGrep {
			grepTool = &tools[i]
			break
		}
	}

	require.NotNil(t, grepTool, "context.grep tool should exist")
	assert.Equal(t, ToolContextGrep, grepTool.Name)
	assert.NotEmpty(t, grepTool.Description)

	// Verify schema has pattern field
	var schemaMap map[string]interface{}
	err := json.Unmarshal(grepTool.InputSchema, &schemaMap)
	require.NoError(t, err)

	properties, ok := schemaMap["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "pattern")
	assert.Contains(t, properties, "include")
	assert.Contains(t, properties, "case_insensitive")

	required, ok := schemaMap["required"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, required, "pattern")
}
