# PlanMark Architecture

This document outlines the architecture, data flow, and runtime components of PlanMark.

## System Overview

PlanMark provides a local graphical review surface for terminal AI coding agents. It receives plans, diffs, HTML components, or architecture diagrams, exposes them through an embedded browser canvas, and exports human edits as structured surgical patch packets.

```
       +--------------------------------------------------------+
       |             CLI Coding Agent Session                   |
       |  (Claude Code / Gemini CLI / Codex / OpenCode / Aider) |
       +--------------------------------------------------------+
                                   |
                                   | Agent outputs Plan / Diff / HTML
                                   | Hook triggers: `planmark review`
                                   v
       +--------------------------------------------------------+
       |                    PlanMark CLI                        |
       |          (Single Go Binary, Sub-10ms Startup)          |
       +--------------------------------------------------------+
                                   |
                     +-------------+-------------+
                     |                           |
                     v                           v
     +-------------------------------+   +-------------------------------+
     |      Ephemeral HTTP Server    |   |     Document AST & Parser     |
     | - Localhost random port       |   | - Markdown (goldmark AST)     |
     | - WebSocket live sync         |   | - Git Unified Diff parser     |
     | - Embedded React 19 assets    |   | - HTML sandbox wrapper        |
     +-------------------------------+   +-------------------------------+
                     |                           |
                     +-------------+-------------+
                                   |
                                   v
       +--------------------------------------------------------+
       |               React 19 Review Canvas                   |
       |  - Mode 1: Document & Plan Annotation (Highlights)     |
       |  - Mode 2: Git Diff Code Review (Line Suggestions)     |
       |  - Mode 3: Interactive HTML / Component Sandbox        |
       |  - Mode 4: Diagram Pinning (Mermaid & Architecture)    |
       +--------------------------------------------------------+
                                   |
                                   | User clicks: [Submit Feedback]
                                   v
       +--------------------------------------------------------+
       |           Surgical Feedback Compiler                   |
       |  - Generates Line-Anchored Mutation Blocks             |
       |  - Builds Deterministic Diff Suggestions               |
       |  - Synthesizes Structured Feedback Packet              |
       +--------------------------------------------------------+
                                   |
                                   | Writes formatted output to stdout
                                   | or returns via MCP / WebSocket / IPC
                                   v
       +--------------------------------------------------------+
       |             Agent Executes Approved Revisions          |
       +--------------------------------------------------------+
```

## Core Modules

### 1. CLI Entrypoint (`cmd/planmark`)
Handles CLI arguments, parses flags, decides between one-off ephemeral mode (`review`, `diff`, pipe) and persistent server mode (`inbox`, `mcp`), and connects standard input/output.

### 2. Parsers (`internal/parser`)
- **Markdown Parser (`markdown.go`)**: Uses `goldmark` to convert markdown into HTML while retaining source line numbers for each heading, paragraph, list item, and code fence block.
- **Diff Parser (`diff.go`)**: Parses unified diff headers, file names, hunk ranges (`@@ -a,b +c,d @@`), and generates line-by-line diff models with old/new line numbers.
- **HTML Sandbox (`html_sandbox.go`)**: Wraps user HTML previews in a secure Content Security Policy (CSP) iframe that allows rendering but prevents access to local storage, cookies, or host networking.
- **Mermaid Validator (`mermaid.go`)**: Validates Mermaid graph syntax and extracts component labels.

### 3. Ephemeral Server & Live Sync (`internal/server`)
- Allocates an open random port on `127.0.0.1`.
- Serves embedded single-page application assets from Go's `embed.FS`.
- Opens the system browser automatically via platform-specific routines (Windows `rundll32/cmd`, macOS `open`, Linux `xdg-open`).
- Maintains a WebSocket connection with the browser. When an agent updates a reviewed file on disk, `fsnotify` detects the modification and pushes the updated AST to the client instantly.

### 4. Surgical Feedback Compiler (`internal/format`)
Translates visual review items into deterministic text blocks:
- **Replacements**: Line-anchored blocks bounded by `<<<<<<< ORIGINAL`, `======= PROPOSED`, and `>>>>>>> RATIONALE`.
- **Suggestions**: GitHub-style fenced code blocks with `suggestion` language tags and exact target file and line pointers.
- **Visual Pins**: Element coordinates, selector, and human notes for HTML components or Mermaid nodes.

### 5. Multi-Agent Session Manager (`internal/store`)
Tracks parallel review requests from background agents:
- Stores session metadata, artifact payloads, timestamps, and current state (`pending`, `in_review`, `submitted`, `rejected`).
- Allows switching between multiple agent requests in the unified inbox without stopping background agent runs.

### 6. Model Context Protocol Server (`internal/mcp`)
Exposes PlanMark capabilities over stdio JSON-RPC so agent runtimes can trigger reviews as native tool calls:
- `request_human_review`: Submits a plan, diff, or mockup and awaits user feedback.
- `get_review_status`: Queries whether a review is pending or completed.
- `cancel_review`: Aborts an active review session.
