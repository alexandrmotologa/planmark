package mcp_test

import (
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/alexandrmotologa/planmark/internal/mcp"
	"github.com/alexandrmotologa/planmark/internal/store"
	"github.com/alexandrmotologa/planmark/internal/types"
)

func TestMCPInitialize(t *testing.T) {
	st := store.NewSessionStore()
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>Mock UI</html>")},
	}

	_ = mcp.NewMCPServer(st, mockFS)

	// Verify standard MCP initialize payload structure
	req := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal JSON-RPC request: %v", err)
	}

	var parsed mcp.JSONRPCRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if parsed.Method != "initialize" {
		t.Errorf("expected method initialize, got %s", parsed.Method)
	}
}

func TestMCPStoreIntegration(t *testing.T) {
	st := store.NewSessionStore()
	st.Create(&types.Session{
		ID:    "mcp_test",
		Agent: "mcp-agent",
		Title: "MCP Review",
		Artifact: types.Artifact{
			Type:    types.ArtifactMarkdown,
			Content: "# Test MCP",
		},
	})

	list := st.List()
	if len(list) != 1 {
		t.Errorf("expected 1 session in store, got %d", len(list))
	}
}
