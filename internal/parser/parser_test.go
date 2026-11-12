package parser_test

import (
	"testing"

	"github.com/alexandrmotologa/planmark/internal/parser"
)

func TestMarkdownParser(t *testing.T) {
	mdContent := `# Main Architecture

This is paragraph one.

## Section 2: Storage
We use SQLite for local caching.

## Section 3: Dispatcher
Workers run in parallel.`

	p := parser.NewMarkdownParser()
	result, err := p.Parse(mdContent)
	if err != nil {
		t.Fatalf("unexpected error parsing markdown: %v", err)
	}

	if len(result.Lines) < 8 {
		t.Errorf("expected at least 8 lines, got %d", len(result.Lines))
	}

	if len(result.Sections) != 3 {
		t.Errorf("expected 3 sections, got %d", len(result.Sections))
	}

	if result.Sections[0].Title != "Main Architecture" {
		t.Errorf("expected title 'Main Architecture', got '%s'", result.Sections[0].Title)
	}
}

func TestDiffParser(t *testing.T) {
	diffContent := `diff --git a/auth.go b/auth.go
index 1234567..89abcdef 100644
--- a/auth.go
+++ b/auth.go
@@ -10,3 +10,4 @@ func Login() {
-    oldAuth()
+    ctx := context.Background()
+    newAuth(ctx)
 }`

	p := parser.NewDiffParser()
	result, err := p.Parse(diffContent)
	if err != nil {
		t.Fatalf("unexpected error parsing diff: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	f := result.Files[0]
	if f.OldPath != "auth.go" || f.NewPath != "auth.go" {
		t.Errorf("expected paths auth.go, got old=%s new=%s", f.OldPath, f.NewPath)
	}

	if f.Additions != 2 || f.Deletions != 1 {
		t.Errorf("expected 2 additions and 1 deletion, got +%d -%d", f.Additions, f.Deletions)
	}
}

func TestMermaidParser(t *testing.T) {
	mmdContent := `sequenceDiagram
    autonumber
    participant Client as Web Browser
    participant Server as Go Daemon
    Client->>Server: WebSocket Connect
`

	p := parser.NewMermaidParser()
	result, err := p.Parse(mmdContent)
	if err != nil {
		t.Fatalf("unexpected error parsing mermaid: %v", err)
	}

	if result.Type != parser.MermaidSequence {
		t.Errorf("expected sequenceDiagram type, got %s", result.Type)
	}

	if len(result.Nodes) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(result.Nodes))
	}

	if result.Nodes[0].ID != "Client" || result.Nodes[0].Label != "Web Browser" {
		t.Errorf("expected Client / Web Browser, got %+v", result.Nodes[0])
	}
}

func TestHTMLSandbox(t *testing.T) {
	rawHTML := `<div class="card">Hello</div>`
	sb := parser.NewHTMLSandbox()
	wrapped := sb.EnsureDocument(rawHTML)

	if !parserContains(wrapped, "__planmark_hover_overlay") {
		t.Errorf("expected wrapped HTML to contain inspector script")
	}
}

func parserContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
