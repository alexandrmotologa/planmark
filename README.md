# PlanMark

Local-first visual review, spec annotation, and surgical patch feedback surface for AI coding agents.

PlanMark runs as a standalone Go binary with an embedded React review interface. It takes markdown plans, unified git diffs, HTML mockups, or Mermaid architecture diagrams produced by terminal AI coding agents (Claude Code, Gemini CLI, OpenAI Codex CLI, Aider, OpenCode, Antigravity), lets developers review and annotate them visually in the browser, and formats the review notes into deterministic surgical patches that agents apply without guessing line numbers or intent.

## Why PlanMark

Terminal coding agents generate detailed plans and multi-file patches, but reviewing them in an 80-column terminal window often leads to vague feedback. PlanMark gives engineers a local graphical surface to review artifacts and converts feedback into machine-actionable patch directives.

- Single standalone executable written in Go with low memory consumption (under 15 MB RAM) and startup time under 10 milliseconds.
- Deterministic surgical feedback blocks (`ORIGINAL`, `PROPOSED`, `RATIONALE`) and line-anchored code suggestions instead of unstructured text prompts.
- Live File Watcher (`fsnotify`) with automatic WebSocket synchronization: detects disk modifications by background agents and refreshes the canvas in real time.
- Direct "Apply Locally & Approve" atomic mutation engine: rewrite source files directly on disk to save agent token consumption.
- Multi-artifact support covering Markdown specification plans, unified git diffs, sandboxed HTML component previews, and interactive Mermaid diagrams.
- Interactive Markdown task checklists (`- [ ]`, `- [x]`) and word-level semantic diff previews.
- Voice dictation via native Web Speech API and agent guardrail presets (`Single-Binary`, `Context & Timeouts`, `TDD First`).
- Multi-agent session inbox to handle parallel agent review requests asynchronously.
- Universal integration through CLI pipe mode, repository hooks, and Model Context Protocol (MCP) endpoints.

## Installation

### From Source (Go)

```bash
go install github.com/alexandrmotologa/planmark/cmd/planmark@latest
```

### Precompiled Binaries

Download the appropriate binary for Linux, macOS, or Windows from the GitHub Releases page.

## Quick Start

### 1. Review a Plan or Spec

```bash
planmark review docs/PLAN.md
```

This launches a local HTTP server on an ephemeral port, opens your default browser, and streams updates over WebSockets. Once you finish annotating, clicking "Submit Feedback" writes the structured patch to stdout or your specified output target and terminates the ephemeral server.

### 2. Review Git Changes

```bash
# Review uncommitted changes
planmark diff

# Review specific branches or commits
planmark diff main..feature-branch
```

### 3. Unix Pipe Mode

```bash
cat draft_plan.md | planmark --pipe | claude
```

### 4. Multi-Agent Inbox

```bash
planmark inbox
```

Runs a persistent local review inbox where background agents post review requests through HTTP or Unix socket endpoints.

## Keyboard Shortcuts

Press `?` in the browser canvas to view all shortcuts:
- `?`: Toggle keyboard shortcuts modal
- `p`: Toggle surgical patch drawer
- `Ctrl + Enter` / `Cmd + Enter`: Open feedback submit and dispatch dialog
- `Esc`: Close any active modal or drawer
- `1 / 2 / 3`: Switch device viewports in HTML mockup mode
- `u / s`: Toggle Unified vs. Split view in Git Diff mode

## Surgical Patch Format

PlanMark translates human visual highlights and gutter suggestions into deterministic blocks:

```markdown
<!-- planmark:feedback-packet v1 -->
# Human Review Feedback on `docs/PLAN.md`

## Annotation 1: Text Replacement (Lines 42-45)
<<<<<<< ORIGINAL
We will use a central Redis cluster for session synchronization.
======= PROPOSED
We will use SQLite in WAL mode with local mutexes to keep deployments single-binary.
>>>>>>> RATIONALE
Avoid extra infrastructure dependencies.

## Annotation 2: Code Suggestion (internal/auth/token.go:78)
```suggestion
func VerifyToken(ctx context.Context, tokenStr string) (*Claims, error) {
    if tokenStr == "" {
        return nil, ErrEmptyToken
    }
    return parseJWT(ctx, tokenStr)
}
```
*Comment: Pass context down and return typed error for empty tokens.*
<!-- planmark:end-packet -->
```

## Agent Integrations

### Claude Code

Install the hook into your repository:

```bash
planmark hook install --agent claude-code
```

### Aider

```bash
git diff | planmark --diff | aider --stdin
```

### Antigravity & IDEs

PlanMark supports local JSON-RPC and Model Context Protocol (MCP) servers:

```bash
planmark mcp
```

## Development

### Prerequisites

- Go 1.22 or higher
- Node.js 20 or higher
- npm 10 or higher

### Build

```bash
# Build frontend and compile single binary
make build

# Run Go tests
make test
```

## License

MIT License. Copyright (c) 2026 Alexandr Motologa.
