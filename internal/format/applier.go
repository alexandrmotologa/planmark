package format

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/alexandrmotologa/planmark/internal/types"
)

type FileApplier struct{}

func NewFileApplier() *FileApplier {
	return &FileApplier{}
}

// ApplyToContent applies replacements and deletions in-memory to the given content string.
func (a *FileApplier) ApplyToContent(content string, annotations []types.Annotation) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Filter and sort applicable annotations descending by startLine (bottom-to-top)
	var validAnns []types.Annotation
	for _, ann := range annotations {
		if (ann.Type == types.AnnotationReplacement || ann.Type == types.AnnotationDeletion) &&
			ann.StartLine > 0 && ann.EndLine >= ann.StartLine {
			validAnns = append(validAnns, ann)
		}
	}

	sort.Slice(validAnns, func(i, j int) bool {
		return validAnns[i].StartLine > validAnns[j].StartLine
	})

	for _, ann := range validAnns {
		startIdx := ann.StartLine - 1
		endIdx := ann.EndLine

		if startIdx >= len(lines) {
			continue
		}
		if endIdx > len(lines) {
			endIdx = len(lines)
		}

		var replacementLines []string
		if ann.Type == types.AnnotationReplacement && strings.TrimSpace(ann.Proposed) != "" {
			replacementLines = strings.Split(strings.TrimRight(ann.Proposed, "\r\n"), "\n")
		}

		// Rebuild lines slice
		newLines := make([]string, 0, len(lines)-(endIdx-startIdx)+len(replacementLines))
		newLines = append(newLines, lines[:startIdx]...)
		newLines = append(newLines, replacementLines...)
		newLines = append(newLines, lines[endIdx:]...)
		lines = newLines
	}

	return strings.Join(lines, "\n"), nil
}

// ApplyToFile reads the file, applies annotations, and atomically writes it back.
func (a *FileApplier) ApplyToFile(filePath string, annotations []types.Annotation) (string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read file '%s': %w", filePath, err)
	}

	updated, err := a.ApplyToContent(string(contentBytes), annotations)
	if err != nil {
		return "", fmt.Errorf("failed applying modifications: %w", err)
	}

	// Atomically write back
	if err := os.WriteFile(filePath, []byte(updated), 0644); err != nil {
		return "", fmt.Errorf("failed writing back to '%s': %w", filePath, err)
	}

	return updated, nil
}
