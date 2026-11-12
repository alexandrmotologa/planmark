package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/alexandrmotologa/planmark/internal/types"
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*types.Session
	order    []string
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*types.Session),
		order:    make([]string, 0),
	}
}

func GenerateID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

func (s *SessionStore) Create(session *types.Session) *types.Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session.ID == "" {
		session.ID = GenerateID("ses")
	}
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now
	if session.Status == "" {
		session.Status = types.StatusPending
	}
	if session.Annotations == nil {
		session.Annotations = make([]types.Annotation, 0)
	}

	s.sessions[session.ID] = session
	s.order = append([]string{session.ID}, s.order...) // most recent first
	return session
}

func (s *SessionStore) Get(id string) (*types.Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	// Return a shallow copy
	cp := *sess
	return &cp, true
}

func (s *SessionStore) List() []*types.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*types.Session, 0, len(s.order))
	for _, id := range s.order {
		if sess, ok := s.sessions[id]; ok {
			cp := *sess
			result = append(result, &cp)
		}
	}
	return result
}

func (s *SessionStore) Update(session *types.Session) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[session.ID]; !ok {
		return false
	}
	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session
	return true
}

func (s *SessionStore) AddAnnotation(sessionID string, ann types.Annotation) (types.Annotation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return ann, false
	}

	if ann.ID == "" {
		ann.ID = GenerateID("ann")
	}
	if ann.CreatedAt.IsZero() {
		ann.CreatedAt = time.Now()
	}

	sess.Annotations = append(sess.Annotations, ann)
	sess.UpdatedAt = time.Now()
	return ann, true
}

func (s *SessionStore) DeleteAnnotation(sessionID, annotationID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return false
	}

	filtered := make([]types.Annotation, 0, len(sess.Annotations))
	found := false
	for _, a := range sess.Annotations {
		if a.ID == annotationID {
			found = true
			continue
		}
		filtered = append(filtered, a)
	}

	if found {
		sess.Annotations = filtered
		sess.UpdatedAt = time.Now()
	}
	return found
}

func (s *SessionStore) SetStatus(sessionID string, status types.SessionStatus) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return false
	}

	sess.Status = status
	sess.UpdatedAt = time.Now()
	return true
}
