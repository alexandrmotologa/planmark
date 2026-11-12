package format

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/planmark/internal/types"
)

type UnifiedPatchGenerator struct{}

func NewUnifiedPatchGenerator() *UnifiedPatchGenerator {
	return &UnifiedPatchGenerator{}
}

// GeneratePatch produces a standard git-applyable unified diff for code replacements
func (g *UnifiedPatchGenerator) GeneratePatch(filePath string, originalLines []string, annotations []types.Annotation) string {
	if len(annotations) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
	sb.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

	for _, ann := range annotations {
		if ann.Type != types.AnnotationReplacement && ann.Type != types.AnnotationDeletion {
			continue
		}

		start := ann.StartLine
		origCount := (ann.EndLine - ann.StartLine) + 1
		if origCount < 1 {
			origCount = 1
		}

		propLines := []string{}
		if ann.Type == types.AnnotationReplacement && strings.TrimSpace(ann.Proposed) != "" {
			propLines = strings.Split(strings.TrimRight(ann.Proposed, "\n"), "\n")
		}
		newCount := len(propLines)

		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", start, origCount, start, newCount))

		origSplit := strings.Split(strings.TrimRight(ann.Original, "\n"), "\n")
		for _, l := range origSplit {
			sb.WriteString("-" + l + "\n")
		}
		for _, l := range propLines {
			sb.WriteString("+" + l + "\n")
		}
	}

	return sb.String()
}
