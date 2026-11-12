package parser

import (
	"bufio"
	"regexp"
	"strings"
)

type MermaidDiagramType string

const (
	MermaidFlowchart MermaidDiagramType = "flowchart"
	MermaidSequence  MermaidDiagramType = "sequenceDiagram"
	MermaidClass     MermaidDiagramType = "classDiagram"
	MermaidState     MermaidDiagramType = "stateDiagram"
	MermaidER        MermaidDiagramType = "erDiagram"
	MermaidUnknown   MermaidDiagramType = "unknown"
)

type MermaidNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Line  int    `json:"line"`
}

type ParsedMermaid struct {
	Type   MermaidDiagramType `json:"type"`
	Nodes  []MermaidNode      `json:"nodes"`
	Source string             `json:"source"`
}

type MermaidParser struct{}

func NewMermaidParser() *MermaidParser {
	return &MermaidParser{}
}

func (p *MermaidParser) Parse(source string) (*ParsedMermaid, error) {
	cleanSource := strings.TrimSpace(source)
	diagType := p.detectType(cleanSource)
	nodes := p.extractNodes(cleanSource, diagType)

	return &ParsedMermaid{
		Type:   diagType,
		Nodes:  nodes,
		Source: cleanSource,
	}, nil
}

func (p *MermaidParser) detectType(source string) MermaidDiagramType {
	firstLine := strings.ToLower(strings.TrimSpace(strings.Split(source, "\n")[0]))
	switch {
	case strings.HasPrefix(firstLine, "flowchart") || strings.HasPrefix(firstLine, "graph"):
		return MermaidFlowchart
	case strings.HasPrefix(firstLine, "sequencediagram"):
		return MermaidSequence
	case strings.HasPrefix(firstLine, "classdiagram"):
		return MermaidClass
	case strings.HasPrefix(firstLine, "statediagram"):
		return MermaidState
	case strings.HasPrefix(firstLine, "erdiagram"):
		return MermaidER
	default:
		return MermaidUnknown
	}
}

var (
	flowchartNodeRegex = regexp.MustCompile(`([a-zA-Z0-9_-]+)\[(.*?)\]|([a-zA-Z0-9_-]+)\((.*?)\)|([a-zA-Z0-9_-]+)\{(.*?)\}`)
	sequenceParticipantRegex = regexp.MustCompile(`(?:participant|actor)\s+([a-zA-Z0-9_-]+)(?:\s+as\s+(.*))?`)
)

func (p *MermaidParser) extractNodes(source string, diagType MermaidDiagramType) []MermaidNode {
	scanner := bufio.NewScanner(strings.NewReader(source))
	var nodes []MermaidNode
	seen := make(map[string]bool)
	lineNum := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if diagType == MermaidSequence {
			matches := sequenceParticipantRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				id := matches[1]
				label := id
				if len(matches) > 2 && matches[2] != "" {
					label = matches[2]
				}
				if !seen[id] {
					seen[id] = true
					nodes = append(nodes, MermaidNode{
						ID:    id,
						Label: label,
						Line:  lineNum,
					})
				}
			}
		} else {
			matches := flowchartNodeRegex.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				id := ""
				label := ""
				if m[1] != "" {
					id = m[1]
					label = m[2]
				} else if m[3] != "" {
					id = m[3]
					label = m[4]
				} else if m[5] != "" {
					id = m[5]
					label = m[6]
				}
				if id != "" && !seen[id] {
					seen[id] = true
					nodes = append(nodes, MermaidNode{
						ID:    id,
						Label: label,
						Line:  lineNum,
					})
				}
			}
		}
		lineNum++
	}

	return nodes
}
