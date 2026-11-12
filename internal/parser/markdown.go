package parser

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MarkdownLine struct {
	Number  int    `json:"number"`
	Content string `json:"content"`
}

type MarkdownSection struct {
	Level     int    `json:"level"`
	Title     string `json:"title"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

type ParsedMarkdown struct {
	Lines    []MarkdownLine    `json:"lines"`
	Sections []MarkdownSection `json:"sections"`
	HTML     string            `json:"html"`
}

type MarkdownParser struct {
	gm goldmark.Markdown
}

func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{
		gm: goldmark.New(),
	}
}

func (p *MarkdownParser) Parse(source string) (*ParsedMarkdown, error) {
	srcBytes := []byte(source)
	lines := p.splitLines(source)

	var buf bytes.Buffer
	if err := p.gm.Convert(srcBytes, &buf); err != nil {
		return nil, err
	}

	doc := p.gm.Parser().Parse(text.NewReader(srcBytes))
	sections := p.extractSections(doc, srcBytes, lines)

	return &ParsedMarkdown{
		Lines:    lines,
		Sections: sections,
		HTML:     buf.String(),
	}, nil
}

func (p *MarkdownParser) splitLines(source string) []MarkdownLine {
	scanner := bufio.NewScanner(strings.NewReader(source))
	var lines []MarkdownLine
	lineNum := 1
	for scanner.Scan() {
		lines = append(lines, MarkdownLine{
			Number:  lineNum,
			Content: scanner.Text(),
		})
		lineNum++
	}
	return lines
}

func (p *MarkdownParser) extractSections(node ast.Node, source []byte, lines []MarkdownLine) []MarkdownSection {
	var sections []MarkdownSection

	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			var headingText strings.Builder
			for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
				if t, ok := child.(*ast.Text); ok {
					headingText.Write(t.Segment.Value(source))
				}
			}

			// Estimate line number by finding match in line list
			title := strings.TrimSpace(headingText.String())
			startLine := 1
			for _, l := range lines {
				if strings.Contains(l.Content, title) {
					startLine = l.Number
					break
				}
			}

			sections = append(sections, MarkdownSection{
				Level:     heading.Level,
				Title:     title,
				StartLine: startLine,
				EndLine:   startLine,
			})
		}
		return ast.WalkContinue, nil
	})

	// Adjust EndLine for each section based on the next section's start
	for i := 0; i < len(sections); i++ {
		if i+1 < len(sections) {
			sections[i].EndLine = sections[i+1].StartLine - 1
		} else {
			sections[i].EndLine = len(lines)
		}
		if sections[i].EndLine < sections[i].StartLine {
			sections[i].EndLine = sections[i].StartLine
		}
	}

	return sections
}
