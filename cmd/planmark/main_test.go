package main

import (
	"testing"

	"github.com/alexandrmotologa/planmark/internal/types"
)

func TestDetectArtifactType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		content  string
		expected types.ArtifactType
	}{
		{
			name:     "Markdown by extension",
			filePath: "docs/PLAN.md",
			content:  "# Title",
			expected: types.ArtifactMarkdown,
		},
		{
			name:     "Diff by extension",
			filePath: "changes.patch",
			content:  "diff --git a/b b/b",
			expected: types.ArtifactDiff,
		},
		{
			name:     "HTML by content",
			filePath: "unknown",
			content:  "<!DOCTYPE html>\n<html><body>Hi</body></html>",
			expected: types.ArtifactHTML,
		},
		{
			name:     "Mermaid by content",
			filePath: "diagram.txt",
			content:  "sequenceDiagram\nactor A\nparticipant B",
			expected: types.ArtifactMermaid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectArtifactType(tt.filePath, tt.content)
			if got != tt.expected {
				t.Errorf("detectArtifactType() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
