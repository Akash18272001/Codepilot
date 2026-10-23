package tools

import (
	"errors"
	"fmt"
	"strings"
)

type ToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

type Tool interface {
	Execute(args map[string]interface{}) (string, error)
}

var registry = make(map[string]Tool)

func Register(name string, t Tool) {
	registry[name] = t
}

func ExecuteTool(req ToolRequest) ToolResult {
	tool, exists := registry[req.Name]
	if !exists {
		return ToolResult{Error: fmt.Sprintf("tool %s not found", req.Name)}
	}

	if err := validateToolCall(req); err != nil {
		return ToolResult{Error: fmt.Sprintf("validation security check failed: %v", err)}
	}

	output, err := tool.Execute(req.Arguments)
	if err != nil {
		return ToolResult{Error: err.Error()}
	}

	return ToolResult{Output: output}
}

func validateToolCall(req ToolRequest) error {
	if req.Name == "query_database" {
		query := fmt.Sprintf("%v", req.Arguments["query"])
		forbidden := []string{"DROP", "DELETE", "TRUNCATE", "ALTER"}
		for _, kw := range forbidden {
			if strings.Contains(strings.ToUpper(query), kw) {
				return errors.New("forbidden SQL command: " + kw)
			}
		}
	}
	return nil
}

type LogSearchTool struct{}
func (t *LogSearchTool) Execute(args map[string]interface{}) (string, error) {
	return "[LOGS] Found 14 HTTP 500 error traces at /api/orders: payload batch size 52 exceeds slice capacity of 50", nil
}

type DatabaseTool struct{}
func (t *DatabaseTool) Execute(args map[string]interface{}) (string, error) {
	return "[DB] Order table state healthy. Max items column limit = 50 items per bulk query transaction.", nil
}

type CodeSearchTool struct{}
func (t *CodeSearchTool) Execute(args map[string]interface{}) (string, error) {
	return "[CODE] Found handler at controllers/orders.go: line 42 throws Panic on len(items) > 50", nil
}

func InitTools() {
	Register("search_logs", &LogSearchTool{})
	Register("query_database", &DatabaseTool{})
	Register("search_code", &CodeSearchTool{})
}
