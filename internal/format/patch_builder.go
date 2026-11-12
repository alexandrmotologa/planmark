package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/planmark/internal/types"
)

type FeedbackCompiler struct{}

func NewFeedbackCompiler() *FeedbackCompiler {
	return &FeedbackCompiler{}
}

func (c *FeedbackCompiler) CompileMarkdown(session *types.Session, submission *types.FeedbackSubmission) string {
	var sb strings.Builder

	sb.WriteString("<!-- planmark:feedback-packet v1 -->\n")
	sb.WriteString(fmt.Sprintf("# Human Review Feedback: %s\n", session.Title))
	sb.WriteString(fmt.Sprintf("Session: %s\n", session.ID))
	sb.WriteString(fmt.Sprintf("Agent: %s\n", session.Agent))
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Status: %s\n", submission.Status))

	if submission.Summary != "" {
		sb.WriteString(fmt.Sprintf("Summary: %s\n", submission.Summary))
	}
	sb.WriteString("\n")

	for i, ann := range submission.Annotations {
		idx := i + 1
		switch ann.Type {
		case types.AnnotationReplacement:
			sb.WriteString(fmt.Sprintf("## Annotation %d: Text Replacement (Lines %d-%d)\n", idx, ann.StartLine, ann.EndLine))
			sb.WriteString("<<<<<<< ORIGINAL\n")
			sb.WriteString(strings.TrimRight(ann.Original, "\n"))
			sb.WriteString("\n=======\n")
			sb.WriteString(strings.TrimRight(ann.Proposed, "\n"))
			sb.WriteString("\n>>>>>>> RATIONALE\n")
			sb.WriteString(strings.TrimSpace(ann.Rationale))
			sb.WriteString("\n\n")

		case types.AnnotationSuggestion:
			target := ann.Target
			if target == "" {
				target = session.Artifact.FilePath
			}
			if target == "" {
				target = "snippet"
			}
			sb.WriteString(fmt.Sprintf("## Annotation %d: Code Suggestion (%s:%d)\n", idx, target, ann.StartLine))
			sb.WriteString("```suggestion\n")
			sb.WriteString(strings.TrimRight(ann.Proposed, "\n"))
			sb.WriteString("\n```\n")
			if ann.Comment != "" {
				sb.WriteString(fmt.Sprintf("*Comment: %s*\n", strings.TrimSpace(ann.Comment)))
			}
			sb.WriteString("\n")

		case types.AnnotationDeletion:
			sb.WriteString(fmt.Sprintf("## Annotation %d: Deletion (Lines %d-%d)\n", idx, ann.StartLine, ann.EndLine))
			sb.WriteString("<<<<<<< ORIGINAL\n")
			sb.WriteString(strings.TrimRight(ann.Original, "\n"))
			sb.WriteString("\n======= DELETION\n")
			sb.WriteString(">>>>>>> RATIONALE\n")
			sb.WriteString(strings.TrimSpace(ann.Rationale))
			sb.WriteString("\n\n")

		case types.AnnotationComment:
			target := ann.Target
			if target == "" && ann.StartLine > 0 {
				target = fmt.Sprintf("Line %d", ann.StartLine)
			}
			sb.WriteString(fmt.Sprintf("## Annotation %d: Comment (%s)\n", idx, target))
			sb.WriteString(fmt.Sprintf("> %s\n\n", strings.TrimSpace(ann.Comment)))

		case types.AnnotationPin:
			sb.WriteString(fmt.Sprintf("## Annotation %d: Visual Pin (%s)\n", idx, ann.Target))
			if ann.Coordinates != nil {
				sb.WriteString(fmt.Sprintf("- Target: `%s`\n", ann.Coordinates.Selector))
				sb.WriteString(fmt.Sprintf("- Coordinates: (x: %.1f, y: %.1f, viewport: %dpx)\n", ann.Coordinates.X, ann.Coordinates.Y, ann.Coordinates.Viewport))
			} else {
				sb.WriteString(fmt.Sprintf("- Target: `%s`\n", ann.Target))
			}
			sb.WriteString(fmt.Sprintf("- Note: \"%s\"\n\n", strings.TrimSpace(ann.Comment)))
		}
	}

	sb.WriteString("<!-- planmark:end-packet -->\n")
	return sb.String()
}

func (c *FeedbackCompiler) CompilePacket(session *types.Session, submission *types.FeedbackSubmission) *types.FeedbackPacket {
	rawMd := c.CompileMarkdown(session, submission)
	return &types.FeedbackPacket{
		Version:     "1.0",
		SessionID:   session.ID,
		Artifact:    session.Artifact.FilePath,
		Status:      submission.Status,
		Summary:     submission.Summary,
		Annotations: submission.Annotations,
		RawMarkdown: rawMd,
	}
}
