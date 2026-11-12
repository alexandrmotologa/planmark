# Surgical Patch Protocol Specification (v1)

## Overview

PlanMark outputs human review notes as structured packets. Rather than emitting conversational text, PlanMark formats each annotation with exact line references, original snippets, proposed alternatives, and explanatory rationale. This enables AI coding agents to execute edits without ambiguity.

## Packet Envelope

A feedback packet begins with `<!-- planmark:feedback-packet v1 -->` and closes with `<!-- planmark:end-packet -->`.

```markdown
<!-- planmark:feedback-packet v1 -->
# Human Review Feedback: <Target File or Artifact>
Session: <Session ID>
Timestamp: <ISO-8601 UTC>
Status: <approved | changes_requested | rejected>

<Annotations>

<!-- planmark:end-packet -->
```

## Annotation Types

### 1. Text Replacement Block

Used for document sections, architectural specifications, or markdown plans where specific paragraphs need revision.

Format:
```markdown
## Annotation <N>: Text Replacement (Lines <Start>-<End>)
<<<<<<< ORIGINAL
<Exact original text from document>
======= PROPOSED
<User proposed replacement text>
>>>>>>> RATIONALE
<Reasoning or instruction provided by reviewer>
```

### 2. Code Suggestion Block

Used in Git diff and code review modes. Modeled after GitHub pull request suggestions so LLMs recognize the format natively.

Format:
```markdown
## Annotation <N>: Code Suggestion (<File Path>:<Line Number>)
```suggestion
<Proposed replacement code lines>
```
*Comment: <Reviewer explanation or note>*
```

### 3. Deletion / Redline Block

Used when a section or block of code should be removed entirely.

Format:
```markdown
## Annotation <N>: Deletion (Lines <Start>-<End>)
<<<<<<< ORIGINAL
<Original lines to delete>
======= DELETION
>>>>>>> RATIONALE
<Reason for deletion>
```

### 4. Visual Pin / Coordinate Annotation

Used for HTML component mockups and Mermaid architecture diagrams.

Format:
```markdown
## Annotation <N>: Visual Pin (<Component Selector or Node ID>)
- Target: `<DOM Selector or Mermaid Node>`
- Coordinates: (x: <X>, y: <Y>, viewport: <390|768|1440>)
- Note: "<Reviewer comment>"
```

## JSON Representation

When requested via `--json` flag or HTTP API, PlanMark returns an identical machine structure:

```json
{
  "version": "1.0",
  "sessionId": "ses_01j7xyz",
  "artifact": "docs/PLAN.md",
  "status": "changes_requested",
  "annotations": [
    {
      "id": "ann_1",
      "type": "replacement",
      "startLine": 42,
      "endLine": 45,
      "original": "We will use Redis...",
      "proposed": "We will use SQLite...",
      "rationale": "Avoid extra infrastructure dependencies."
    },
    {
      "id": "ann_2",
      "type": "suggestion",
      "file": "internal/auth/token.go",
      "line": 78,
      "proposed": "func VerifyToken(ctx context.Context, tokenStr string) (*Claims, error) {\n...",
      "comment": "Pass context down and return typed error."
    }
  ]
}
```
