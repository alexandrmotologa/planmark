package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/alexandrmotologa/planmark/internal/parser"
	"github.com/alexandrmotologa/planmark/internal/server"
	"github.com/alexandrmotologa/planmark/internal/store"
	"github.com/alexandrmotologa/planmark/internal/types"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type MCPServer struct {
	store  *store.SessionStore
	assets fs.FS
}

func NewMCPServer(st *store.SessionStore, assets fs.FS) *MCPServer {
	return &MCPServer{
		store:  st,
		assets: assets,
	}
}

func (s *MCPServer) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(req.ID, -32700, "Parse error")
			continue
		}

		s.handleRequest(&req)
	}
	return scanner.Err()
}

func (s *MCPServer) handleRequest(req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.sendResult(req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]string{
				"name":    "planmark",
				"version": "1.0.0",
			},
		})

	case "notifications/initialized":
		// No response needed for notification

	case "tools/list":
		s.sendResult(req.ID, map[string]any{
			"tools": []map[string]any{
				{
					"name":        "request_human_review",
					"description": "Launches a local-first PlanMark review session in the browser and blocks until the human developer approves or submits surgical patch feedback.",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"title": map[string]string{
								"type":        "string",
								"description": "Short title describing the review request",
							},
							"type": map[string]any{
								"type":        "string",
								"enum":        []string{"markdown", "diff", "html", "mermaid"},
								"description": "Artifact type",
							},
							"content": map[string]string{
								"type":        "string",
								"description": "The raw markdown, unified diff, HTML, or Mermaid diagram source",
							},
							"filePath": map[string]string{
								"type":        "string",
								"description": "Original file path if applicable",
							},
						},
						"required": []string{"title", "type", "content"},
					},
				},
				{
					"name":        "get_active_sessions",
					"description": "Returns list of all active PlanMark review sessions.",
					"inputSchema": map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				},
			},
		})

	case "tools/call":
		var callParams struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			s.sendError(req.ID, -32602, "Invalid params")
			return
		}

		if callParams.Name == "request_human_review" {
			s.executeReviewTool(req.ID, callParams.Arguments)
		} else if callParams.Name == "get_active_sessions" {
			sessions := s.store.List()
			s.sendResult(req.ID, map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": fmt.Sprintf("%d active review sessions", len(sessions)),
					},
				},
			})
		} else {
			s.sendError(req.ID, -32601, fmt.Sprintf("Unknown tool: %s", callParams.Name))
		}

	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *MCPServer) executeReviewTool(reqID any, args map[string]any) {
	title, _ := args["title"].(string)
	artTypeStr, _ := args["type"].(string)
	content, _ := args["content"].(string)
	filePath, _ := args["filePath"].(string)

	artType := types.ArtifactType(artTypeStr)
	if artType == "" {
		artType = types.ArtifactMarkdown
	}

	session := &types.Session{
		Agent: "mcp-agent",
		Title: title,
		Artifact: types.Artifact{
			Type:     artType,
			Title:    title,
			FilePath: filePath,
			Content:  content,
		},
	}

	// Enrich artifact
	switch artType {
	case types.ArtifactMarkdown:
		if p, err := parser.NewMarkdownParser().Parse(content); err == nil {
			session.Artifact.Parsed = p
		}
	case types.ArtifactDiff:
		if p, err := parser.NewDiffParser().Parse(content); err == nil {
			session.Artifact.Parsed = p
		}
	case types.ArtifactMermaid:
		if p, err := parser.NewMermaidParser().Parse(content); err == nil {
			session.Artifact.Parsed = p
		}
	}

	created := s.store.Create(session)
	port, err := server.GetFreePort()
	if err != nil {
		s.sendError(reqID, -32000, fmt.Sprintf("failed to get open port: %v", err))
		return
	}

	doneChan := make(chan *types.FeedbackPacket, 1)
	srv := server.NewServer(server.ServerConfig{
		Port:         port,
		Store:        s.store,
		Assets:       s.assets,
		DoneChan:     doneChan,
		AutoShutdown: true,
	})

	go func() {
		_ = srv.Start()
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/session/%s", port, created.ID)
	_ = server.OpenBrowser(url)

	// Block until user submits feedback or timeout (30 minutes)
	select {
	case packet := <-doneChan:
		_ = srv.Shutdown()
		s.sendResult(reqID, map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": packet.RawMarkdown,
				},
			},
		})
	case <-time.After(30 * time.Minute):
		_ = srv.Shutdown()
		s.sendError(reqID, -32000, "Review session timed out after 30 minutes")
	}
}

func (s *MCPServer) sendResult(id any, result any) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func (s *MCPServer) sendError(id any, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]any{
			"code":    code,
			"message": message,
		},
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}
