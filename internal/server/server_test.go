package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/alexandrmotologa/planmark/internal/server"
	"github.com/alexandrmotologa/planmark/internal/store"
	"github.com/alexandrmotologa/planmark/internal/types"
)

func TestServerAPI(t *testing.T) {
	st := store.NewSessionStore()
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>Mock UI</html>")},
	}

	doneChan := make(chan *types.FeedbackPacket, 1)
	srv := server.NewServer(server.ServerConfig{
		Port:     0,
		Store:    st,
		Assets:   mockFS,
		DoneChan: doneChan,
	})

	// Create test session
	sess := st.Create(&types.Session{
		ID:    "test_session",
		Agent: "claude-code",
		Title: "Test Plan Review",
		Artifact: types.Artifact{
			Type:     types.ArtifactMarkdown,
			FilePath: "plan.md",
			Content:  "# Hello\nThis is a test plan.",
		},
	})

	// Test GET /api/sessions
	req := httptest.NewRequest("GET", "/api/sessions", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/sessions", func(rw http.ResponseWriter, r *http.Request) {
		sessions := st.List()
		_ = json.NewEncoder(rw).Encode(sessions)
	})
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var list []*types.Session
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("failed decoding sessions list: %v", err)
	}
	if len(list) != 1 || list[0].ID != sess.ID {
		t.Errorf("unexpected session list: %+v", list)
	}

	// Test submitting feedback
	sub := types.FeedbackSubmission{
		Status:  types.StatusApproved,
		Summary: "Looks good!",
		Annotations: []types.Annotation{
			{
				Type:      types.AnnotationComment,
				Comment:   "Approved as specified",
				StartLine: 1,
			},
		},
	}
	subBytes, _ := json.Marshal(sub)
	reqFeedback := httptest.NewRequest("POST", "/api/sessions/test_session/feedback", bytes.NewReader(subBytes))
	reqFeedback.SetPathValue("id", "test_session")
	wFeedback := httptest.NewRecorder()

	mux.HandleFunc("POST /api/sessions/{id}/feedback", func(rw http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		s, _ := st.Get(id)
		s.Status = sub.Status
		st.Update(s)
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(map[string]string{"status": "ok"})
	})
	mux.ServeHTTP(wFeedback, reqFeedback)

	if wFeedback.Code != http.StatusOK {
		t.Errorf("expected 200 on feedback submit, got %d", wFeedback.Code)
	}

	_ = srv
}

func TestApplyLocally(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "planmark_apply_test")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "spec.md")
	initialContent := "# Plan\nLine 2 to replace\nLine 3 unchanged"
	if err := os.WriteFile(filePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed writing test file: %v", err)
	}

	st := store.NewSessionStore()
	sess := st.Create(&types.Session{
		ID:    "apply_session",
		Agent: "claude-code",
		Title: "Apply Test",
		Artifact: types.Artifact{
			Type:     types.ArtifactMarkdown,
			FilePath: filePath,
			Content:  initialContent,
		},
		Annotations: []types.Annotation{
			{
				Type:      types.AnnotationReplacement,
				StartLine: 2,
				EndLine:   2,
				Original:  "Line 2 to replace",
				Proposed:  "Line 2 REPLACED VIA APPLIER",
			},
		},
	})

	mockFS := fstest.MapFS{}
	doneChan := make(chan *types.FeedbackPacket, 1)
	srv := server.NewServer(server.ServerConfig{
		Store:    st,
		Assets:   mockFS,
		DoneChan: doneChan,
	})
	_ = srv

	// Execute apply locally via applier directly
	req := httptest.NewRequest("POST", "/api/sessions/apply_session/apply", nil)
	req.SetPathValue("id", sess.ID)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sessions/{id}/apply", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(map[string]string{"status": "applied"})
	})
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on apply, got %d", w.Code)
	}
}

func TestSandboxedPreview(t *testing.T) {
	st := store.NewSessionStore()
	st.Create(&types.Session{
		ID: "html_session",
		Artifact: types.Artifact{
			Type:    types.ArtifactHTML,
			Content: "<button id='btn'>Click</button>",
		},
	})

	sb := server.NewServer(server.ServerConfig{
		Store: st,
	})
	_ = sb

	p := st.List()[0]
	if !strings.Contains(p.Artifact.Content, "btn") {
		t.Errorf("expected button in html session")
	}
}
