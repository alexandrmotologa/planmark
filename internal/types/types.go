package types

import "time"

type ArtifactType string

const (
	ArtifactMarkdown ArtifactType = "markdown"
	ArtifactDiff     ArtifactType = "diff"
	ArtifactHTML     ArtifactType = "html"
	ArtifactMermaid  ArtifactType = "mermaid"
)

type SessionStatus string

const (
	StatusPending          SessionStatus = "pending"
	StatusInReview         SessionStatus = "in_review"
	StatusApproved         SessionStatus = "approved"
	StatusChangesRequested SessionStatus = "changes_requested"
	StatusRejected         SessionStatus = "rejected"
)

type AnnotationType string

const (
	AnnotationReplacement AnnotationType = "replacement"
	AnnotationSuggestion  AnnotationType = "suggestion"
	AnnotationDeletion    AnnotationType = "deletion"
	AnnotationComment     AnnotationType = "comment"
	AnnotationPin         AnnotationType = "pin"
)

type Point struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Viewport int     `json:"viewport"` // 390, 768, 1440
	Selector string  `json:"selector,omitempty"`
}

type Annotation struct {
	ID          string         `json:"id"`
	Type        AnnotationType `json:"type"`
	Target      string         `json:"target,omitempty"` // file path, selector, or node ID
	StartLine   int            `json:"startLine,omitempty"`
	EndLine     int            `json:"endLine,omitempty"`
	Original    string         `json:"original,omitempty"`
	Proposed    string         `json:"proposed,omitempty"`
	Rationale   string         `json:"rationale,omitempty"`
	Comment     string         `json:"comment,omitempty"`
	Coordinates *Point         `json:"coordinates,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type Artifact struct {
	Type     ArtifactType `json:"type"`
	Title    string       `json:"title"`
	FilePath string       `json:"filePath,omitempty"`
	Content  string       `json:"content"`
	Parsed   any          `json:"parsed,omitempty"`
}

type Session struct {
	ID          string        `json:"id"`
	Agent       string        `json:"agent"`
	Title       string        `json:"title"`
	Artifact    Artifact      `json:"artifact"`
	Status      SessionStatus `json:"status"`
	Annotations []Annotation  `json:"annotations"`
	Summary     string        `json:"summary,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

type FeedbackSubmission struct {
	Status      SessionStatus `json:"status"`
	Summary     string        `json:"summary,omitempty"`
	Annotations []Annotation  `json:"annotations"`
}

type FeedbackPacket struct {
	Version     string        `json:"version"`
	SessionID   string        `json:"sessionId"`
	Artifact    string        `json:"artifact"`
	Status      SessionStatus `json:"status"`
	Summary     string        `json:"summary,omitempty"`
	Annotations []Annotation  `json:"annotations"`
	RawMarkdown string        `json:"rawMarkdown"`
}
