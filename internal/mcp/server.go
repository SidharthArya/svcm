package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"lsysctl/internal/core"
	"os"
)

// Minimal JSON-RPC 2.0 types for MCP
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	InputSchemaSchema interface{} `json:"inputSchema"`
}

// MCP Server
func Run() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		handleRequest(line)
	}
}

func handleRequest(data []byte) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return // Ignore malformed
	}

	var res interface{}
	var err *JSONRPCError

	switch req.Method {
	case "initialize":
		res = map[string]interface{}{
			"protocolVersion": "0.1.0",
			"serverInfo": map[string]string{
				"name":    "lsysctl-mcp",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		}
	case "notifications/initialized":
		// No response needed
		return
	case "tools/list":
		res = map[string]interface{}{
			"tools": []Tool{
				{
					Name:        "list_services",
					Description: "List all systemd services for the user",
					InputSchemaSchema: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					Name:        "start_service",
					Description: "Start a systemd service",
					InputSchemaSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]string{"type": "string"},
						},
						"required": []string{"name"},
					},
				},
				{
					Name:        "stop_service",
					Description: "Stop a systemd service",
					InputSchemaSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]string{"type": "string"},
						},
						"required": []string{"name"},
					},
				},
			},
		}
	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		json.Unmarshal(req.Params, &params)

		manager, e := core.NewSystemdManager()
		if e != nil {
			err = &JSONRPCError{Code: -32000, Message: e.Error()}
			break
		}
		defer manager.Close()

		switch params.Name {
		case "list_services":
			list, e := manager.ListServices()
			if e != nil {
				err = &JSONRPCError{Code: -32000, Message: e.Error()}
			} else {
				// MCP expects text or artifacts
				txt := ""
				for _, s := range list {
					txt += fmt.Sprintf("%-30s %s %s\n", s.Name, s.ActiveState, s.Description)
				}
				res = map[string]interface{}{
					"content": []map[string]string{
						{"type": "text", "text": txt},
					},
				}
			}
		case "start_service":
			name := params.Arguments["name"].(string)
			if e := manager.StartService(name); e != nil {
				err = &JSONRPCError{Code: -32000, Message: e.Error()}
			} else {
				res = map[string]interface{}{
					"content": []map[string]string{
						{"type": "text", "text": "Service started"},
					},
				}
			}
		case "stop_service":
			name := params.Arguments["name"].(string)
			if e := manager.StopService(name); e != nil {
				err = &JSONRPCError{Code: -32000, Message: e.Error()}
			} else {
				res = map[string]interface{}{
					"content": []map[string]string{
						{"type": "text", "text": "Service stopped"},
					},
				}
			}
		default:
			err = &JSONRPCError{Code: -32601, Message: "Method not found"}
		}
	default:
		// Just ignore unknowns or pings
		return
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  res,
		Error:   err,
		ID:      req.ID,
	}
	bytes, _ := json.Marshal(response)
	fmt.Printf("%s\n", bytes)
}
