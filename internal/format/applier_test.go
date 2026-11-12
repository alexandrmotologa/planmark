package format_test

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/planmark/internal/format"
	"github.com/alexandrmotologa/planmark/internal/types"
)

func TestFileApplier(t *testing.T) {
	origContent := `Line 1: Header
Line 2: Section 1
Line 3: Section 2 to be replaced
Line 4: Section 3 to be deleted
Line 5: Footer`

	annotations := []types.Annotation{
		{
			Type:      types.AnnotationReplacement,
			StartLine: 3,
			EndLine:   3,
			Proposed:  "Line 3: Section 2 REVISED",
		},
		{
			Type:      types.AnnotationDeletion,
			StartLine: 4,
			EndLine:   4,
		},
	}

	applier := format.NewFileApplier()
	updated, err := applier.ApplyToContent(origContent, annotations)
	if err != nil {
		t.Fatalf("unexpected error applying annotations: %v", err)
	}

	lines := strings.Split(updated, "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines after 1 replacement and 1 deletion, got %d. Content:\n%s", len(lines), updated)
	}

	if lines[2] != "Line 3: Section 2 REVISED" {
		t.Errorf("expected line 3 to be revised, got: %s", lines[2])
	}
	if lines[3] != "Line 5: Footer" {
		t.Errorf("expected line 4 to be footer, got: %s", lines[3])
	}
}
