package store_test

import (
	"sync"
	"testing"
	"time"

	"github.com/alexandrmotologa/planmark/internal/store"
	"github.com/alexandrmotologa/planmark/internal/types"
)

func TestSessionStoreOperations(t *testing.T) {
	st := store.NewSessionStore()

	sess := st.Create(&types.Session{
		Agent: "test-agent",
		Title: "Test Spec",
		Artifact: types.Artifact{
			Type:    types.ArtifactMarkdown,
			Content: "# Hello",
		},
	})

	if sess.ID == "" {
		t.Errorf("expected session ID to be generated")
	}

	retrieved, ok := st.Get(sess.ID)
	if !ok {
		t.Fatalf("expected session to exist")
	}
	if retrieved.Title != "Test Spec" {
		t.Errorf("expected title 'Test Spec', got '%s'", retrieved.Title)
	}

	// Test adding annotation
	ann, ok := st.AddAnnotation(sess.ID, types.Annotation{
		Type:      types.AnnotationReplacement,
		StartLine: 1,
		EndLine:   1,
		Original:  "Hello",
		Proposed:  "Greetings",
		Rationale: "Polite phrasing",
	})
	if !ok || ann.ID == "" {
		t.Fatalf("expected annotation to be added with ID")
	}

	// Verify annotation is in session
	updated, _ := st.Get(sess.ID)
	if len(updated.Annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(updated.Annotations))
	}

	// Test deleting annotation
	deleted := st.DeleteAnnotation(sess.ID, ann.ID)
	if !deleted {
		t.Errorf("expected annotation to be deleted")
	}

	updatedAfterDelete, _ := st.Get(sess.ID)
	if len(updatedAfterDelete.Annotations) != 0 {
		t.Errorf("expected 0 annotations after delete, got %d", len(updatedAfterDelete.Annotations))
	}

	// Test setting status
	st.SetStatus(sess.ID, types.StatusApproved)
	finalSess, _ := st.Get(sess.ID)
	if finalSess.Status != types.StatusApproved {
		t.Errorf("expected status 'approved', got '%s'", finalSess.Status)
	}
}

func TestStoreConcurrency(t *testing.T) {
	st := store.NewSessionStore()
	sess := st.Create(&types.Session{
		Agent: "concurrent-agent",
		Title: "Concurrent Session",
	})

	var wg sync.WaitGroup
	workers := 10
	annotationsPerWorker := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < annotationsPerWorker; j++ {
				st.AddAnnotation(sess.ID, types.Annotation{
					Type:      types.AnnotationComment,
					Comment:   "Concurrent comment",
					CreatedAt: time.Now(),
				})
			}
		}(i)
	}

	wg.Wait()

	finalSess, _ := st.Get(sess.ID)
	expectedTotal := workers * annotationsPerWorker
	if len(finalSess.Annotations) != expectedTotal {
		t.Errorf("expected %d annotations, got %d", expectedTotal, len(finalSess.Annotations))
	}
}
