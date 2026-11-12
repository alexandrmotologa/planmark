package format_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/planmark/internal/format"
	"github.com/alexandrmotologa/planmark/internal/types"
)

func TestFeedbackCompiler(t *testing.T) {
	session := &types.Session{
		ID:    "ses_123",
		Agent: "claude-code",
		Title: "Database Refactor",
		Artifact: types.Artifact{
			Type:     types.ArtifactMarkdown,
			FilePath: "docs/PLAN.md",
		},
	}

	submission := &types.FeedbackSubmission{
		Status:  types.StatusChangesRequested,
		Summary: "Please adjust the database engine and token validator.",
		Annotations: []types.Annotation{
			{
				ID:        "ann_1",
				Type:      types.AnnotationReplacement,
				StartLine: 42,
				EndLine:   44,
				Original:  "Use Redis cluster.",
				Proposed:  "Use SQLite in WAL mode.",
				Rationale: "Avoid external daemon.",
				CreatedAt: time.Now(),
			},
			{
				ID:        "ann_2",
				Type:      types.AnnotationSuggestion,
				Target:    "internal/auth/token.go",
				StartLine: 78,
				Proposed:  "func Verify(ctx context.Context) error",
				Comment:   "Propagate context",
				CreatedAt: time.Now(),
			},
		},
	}

	compiler := format.NewFeedbackCompiler()
	md := compiler.CompileMarkdown(session, submission)

	if !strings.Contains(md, "<!-- planmark:feedback-packet v1 -->") {
		t.Errorf("missing packet opening header")
	}
	if !strings.Contains(md, "<<<<<<< ORIGINAL\nUse Redis cluster.\n=======\nUse SQLite in WAL mode.\n>>>>>>> RATIONALE\nAvoid external daemon.") {
		t.Errorf("missing or incorrect replacement block")
	}
	if !strings.Contains(md, "```suggestion\nfunc Verify(ctx context.Context) error\n```") {
		t.Errorf("missing suggestion block")
	}
	if !strings.Contains(md, "<!-- planmark:end-packet -->") {
		t.Errorf("missing packet closing footer")
	}
}
