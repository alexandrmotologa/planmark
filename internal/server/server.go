package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alexandrmotologa/planmark/internal/format"
	"github.com/alexandrmotologa/planmark/internal/parser"
	"github.com/alexandrmotologa/planmark/internal/store"
	"github.com/alexandrmotologa/planmark/internal/types"
	"github.com/gorilla/websocket"
)

type ServerConfig struct {
	Port         int
	Store        *store.SessionStore
	Assets       fs.FS
	DoneChan     chan *types.FeedbackPacket
	Compiler     *format.FeedbackCompiler
	AutoShutdown bool
}

type Server struct {
	cfg        ServerConfig
	httpServer *http.Server
	upgrader   websocket.Upgrader
	clientsMu  sync.Mutex
	clients    map[*websocket.Conn]bool
}

func NewServer(cfg ServerConfig) *Server {
	if cfg.Compiler == nil {
		cfg.Compiler = format.NewFeedbackCompiler()
	}
	s := &Server{
		cfg: cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]bool),
	}
	return s
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/sessions", s.handleListSessions)
	mux.HandleFunc("POST /api/sessions", s.handleCreateSession)
	mux.HandleFunc("GET /api/sessions/{id}", s.handleGetSession)
	mux.HandleFunc("POST /api/sessions/{id}/annotations", s.handleAddAnnotation)
	mux.HandleFunc("DELETE /api/sessions/{id}/annotations/{annId}", s.handleDeleteAnnotation)
	mux.HandleFunc("POST /api/sessions/{id}/feedback", s.handleSubmitFeedback)
	mux.HandleFunc("GET /api/sessions/{id}/sandbox", s.handleSandboxPreview)
	mux.HandleFunc("GET /ws", s.handleWebSocket)

	// Static SPA assets handler
	fileServer := http.FileServer(http.FS(s.cfg.Assets))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try to open from embedded fs
		f, err := s.cfg.Assets.Open(path)
		if err != nil {
			// SPA fallback: serve index.html
			indexFile, err := s.cfg.Assets.Open("index.html")
			if err != nil {
				http.Error(w, "Asset not found", http.StatusNotFound)
				return
			}
			defer indexFile.Close()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.Copy(w, indexFile)
			return
		}
		defer f.Close()
		fileServer.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf("127.0.0.1:%d", s.cfg.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.cfg.Store.List()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sessions)
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var session types.Session
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		http.Error(w, fmt.Sprintf("invalid payload: %v", err), http.StatusBadRequest)
		return
	}

	// Auto-parse artifact if content is provided
	s.enrichArtifact(&session.Artifact)

	created := s.cfg.Store.Create(&session)
	s.broadcast(map[string]any{
		"type":    "SESSION_CREATED",
		"session": created,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	session, ok := s.cfg.Store.Get(id)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(session)
}

func (s *Server) handleAddAnnotation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var ann types.Annotation
	if err := json.NewDecoder(r.Body).Decode(&ann); err != nil {
		http.Error(w, fmt.Sprintf("invalid annotation: %v", err), http.StatusBadRequest)
		return
	}

	createdAnn, ok := s.cfg.Store.AddAnnotation(id, ann)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	s.broadcast(map[string]any{
		"type":         "ANNOTATION_ADDED",
		"sessionId":    id,
		"annotation":   createdAnn,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createdAnn)
}

func (s *Server) handleDeleteAnnotation(w http.ResponseWriter, r *http.Request) {
	sessionId := r.PathValue("id")
	annId := r.PathValue("annId")

	if !s.cfg.Store.DeleteAnnotation(sessionId, annId) {
		http.Error(w, "annotation not found", http.StatusNotFound)
		return
	}

	s.broadcast(map[string]any{
		"type":         "ANNOTATION_DELETED",
		"sessionId":    sessionId,
		"annotationId": annId,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	sessionId := r.PathValue("id")
	session, ok := s.cfg.Store.Get(sessionId)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	var sub types.FeedbackSubmission
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, fmt.Sprintf("invalid feedback submission: %v", err), http.StatusBadRequest)
		return
	}

	// Update session status and annotations if provided
	if len(sub.Annotations) > 0 {
		session.Annotations = sub.Annotations
	}
	session.Status = sub.Status
	session.Summary = sub.Summary
	s.cfg.Store.Update(session)

	// Compile surgical patch packet
	packet := s.cfg.Compiler.CompilePacket(session, &sub)

	s.broadcast(map[string]any{
		"type":      "FEEDBACK_SUBMITTED",
		"sessionId": sessionId,
		"packet":    packet,
	})

	// If a CLI completion channel is waiting, deliver the packet
	if s.cfg.DoneChan != nil {
		go func() {
			time.Sleep(50 * time.Millisecond) // small grace for HTTP response to flush
			s.cfg.DoneChan <- packet
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(packet)
}

func (s *Server) handleSandboxPreview(w http.ResponseWriter, r *http.Request) {
	sessionId := r.PathValue("id")
	session, ok := s.cfg.Store.Get(sessionId)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	sandbox := parser.NewHTMLSandbox()
	for k, v := range sandbox.SecurityHeaders() {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	wrapped := sandbox.EnsureDocument(session.Artifact.Content)
	_, _ = w.Write([]byte(wrapped))
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		_ = conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) broadcast(payload any) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for client := range s.clients {
		_ = client.WriteMessage(websocket.TextMessage, data)
	}
}

func (s *Server) enrichArtifact(art *types.Artifact) {
	switch art.Type {
	case types.ArtifactMarkdown:
		p := parser.NewMarkdownParser()
		if parsed, err := p.Parse(art.Content); err == nil {
			art.Parsed = parsed
		}
	case types.ArtifactDiff:
		p := parser.NewDiffParser()
		if parsed, err := p.Parse(art.Content); err == nil {
			art.Parsed = parsed
		}
	case types.ArtifactMermaid:
		p := parser.NewMermaidParser()
		if parsed, err := p.Parse(art.Content); err == nil {
			art.Parsed = parsed
		}
	}
}
