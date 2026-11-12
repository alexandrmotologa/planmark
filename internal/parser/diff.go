package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type DiffLineType string

const (
	DiffLineContext  DiffLineType = "context"
	DiffLineAddition DiffLineType = "add"
	DiffLineDeletion DiffLineType = "delete"
	DiffLineHeader   DiffLineType = "header"
)

type DiffLine struct {
	Type    DiffLineType `json:"type"`
	Content string       `json:"content"`
	OldLine int          `json:"oldLine,omitempty"`
	NewLine int          `json:"newLine,omitempty"`
}

type DiffHunk struct {
	Header    string     `json:"header"`
	OldStart  int        `json:"oldStart"`
	OldLines  int        `json:"oldLines"`
	NewStart  int        `json:"newStart"`
	NewLines  int        `json:"newLines"`
	Lines     []DiffLine `json:"lines"`
}

type DiffFile struct {
	OldPath   string     `json:"oldPath"`
	NewPath   string     `json:"newPath"`
	IsNew     bool       `json:"isNew"`
	IsDeleted bool       `json:"isDeleted"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
	Hunks     []DiffHunk `json:"hunks"`
}

type ParsedDiff struct {
	Files          []DiffFile `json:"files"`
	TotalAdditions int        `json:"totalAdditions"`
	TotalDeletions int        `json:"totalDeletions"`
}

var hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(.*)$`)

type DiffParser struct{}

func NewDiffParser() *DiffParser {
	return &DiffParser{}
}

func (p *DiffParser) Parse(rawDiff string) (*ParsedDiff, error) {
	scanner := bufio.NewScanner(strings.NewReader(rawDiff))
	result := &ParsedDiff{}

	var currentFile *DiffFile
	var currentHunk *DiffHunk
	oldLineNum := 0
	newLineNum := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "diff --git ") {
			if currentHunk != nil && currentFile != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
				currentHunk = nil
			}
			if currentFile != nil {
				result.Files = append(result.Files, *currentFile)
			}

			parts := strings.Split(line, " ")
			oldPath := ""
			newPath := ""
			if len(parts) >= 4 {
				oldPath = strings.TrimPrefix(parts[2], "a/")
				newPath = strings.TrimPrefix(parts[3], "b/")
			}

			currentFile = &DiffFile{
				OldPath: oldPath,
				NewPath: newPath,
			}
			continue
		}

		if currentFile == nil {
			// Single file diff without git header
			currentFile = &DiffFile{
				OldPath: "diff.patch",
				NewPath: "diff.patch",
			}
		}

		if strings.HasPrefix(line, "new file mode ") {
			currentFile.IsNew = true
			continue
		}
		if strings.HasPrefix(line, "deleted file mode ") {
			currentFile.IsDeleted = true
			continue
		}
		if strings.HasPrefix(line, "--- ") {
			path := strings.TrimPrefix(line, "--- ")
			path = strings.TrimPrefix(path, "a/")
			if path != "/dev/null" && currentFile.OldPath == "" {
				currentFile.OldPath = path
			}
			continue
		}
		if strings.HasPrefix(line, "+++ ") {
			path := strings.TrimPrefix(line, "+++ ")
			path = strings.TrimPrefix(path, "b/")
			if path != "/dev/null" && currentFile.NewPath == "" {
				currentFile.NewPath = path
			}
			continue
		}

		// Check for hunk header
		if strings.HasPrefix(line, "@@ ") {
			if currentHunk != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}

			matches := hunkHeaderRegex.FindStringSubmatch(line)
			if len(matches) < 4 {
				return nil, fmt.Errorf("malformed hunk header: %s", line)
			}

			oldStart, _ := strconv.Atoi(matches[1])
			oldLines := 1
			if matches[2] != "" {
				oldLines, _ = strconv.Atoi(matches[2])
			}

			newStart, _ := strconv.Atoi(matches[3])
			newLines := 1
			if matches[4] != "" {
				newLines, _ = strconv.Atoi(matches[4])
			}

			oldLineNum = oldStart
			newLineNum = newStart

			currentHunk = &DiffHunk{
				Header:   line,
				OldStart: oldStart,
				OldLines: oldLines,
				NewStart: newStart,
				NewLines: newLines,
			}
			continue
		}

		if currentHunk != nil {
			if strings.HasPrefix(line, "+") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    DiffLineAddition,
					Content: line[1:],
					NewLine: newLineNum,
				})
				newLineNum++
				currentFile.Additions++
				result.TotalAdditions++
			} else if strings.HasPrefix(line, "-") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    DiffLineDeletion,
					Content: line[1:],
					OldLine: oldLineNum,
				})
				oldLineNum++
				currentFile.Deletions++
				result.TotalDeletions++
			} else if strings.HasPrefix(line, " ") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    DiffLineContext,
					Content: line[1:],
					OldLine: oldLineNum,
					NewLine: newLineNum,
				})
				oldLineNum++
				newLineNum++
			} else if strings.HasPrefix(line, `\ No newline`) {
				// Ignore newline warning
				continue
			}
		}
	}

	if currentHunk != nil && currentFile != nil {
		currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
	}
	if currentFile != nil {
		result.Files = append(result.Files, *currentFile)
	}

	return result, scanner.Err()
}
