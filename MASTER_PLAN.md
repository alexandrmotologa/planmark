# `planmark` — Architecture & Engineering Master Plan
> **The Local-First Visual Review, Spec Annotation & Structured Feedback Surface for AI Coding Agents**  
> *Target: High-Performance Single-Binary Go Daemon (< 15MB RAM) + Embedded React 19 Studio*

---

## 1. Executive Summary & Market Problem

### 1.1 The Terminal Review Bottleneck in AI Coding
The software engineering workflow is undergoing a monumental paradigm shift driven by terminal-based autonomous coding agents (**Claude Code, Gemini CLI, OpenAI Codex CLI, Aider, OpenCode, and Antigravity**). Developers increasingly delegate high-level features, multi-file refactors, and architecture design directly to AI CLI tools.

However, the interaction model hits a severe wall during the **Review & Decision Phase**:
1. **The Terminal Review Nightmare**: When an agent proposes a 300-line architectural plan (`PLAN.md`), a complex 15-file git diff, or an interactive HTML mockup, reviewing it inside a standard 80-column terminal terminal emulator is exhausting and error-prone.
2. **Ambiguous Natural Language Feedback**: When an engineer wants to critique a plan, they must type clumsy terminal instructions: *"In section 2.3 change the database schema, and also on line 45 of auth.go pass context instead of string."* The agent frequently misinterprets which exact paragraph or line was referenced, leading to recursive hallucination loops.
3. **Plannotator’s Architectural Flaws**: While tools like *Plannotator* proved the demand for browser review, their implementation suffers from heavy Node.js/Bun monorepo dependencies (requiring dozens of npm packages, slow startup times, cookie workarounds for random ports, and unstructured raw-text feedback).

### 1.2 The `planmark` Solution
`planmark` is an ultra-fast, zero-dependency, single-binary Go utility with an embedded React 19 visual annotation studio. It acts as the definitive **Human-in-the-Loop visual review layer** for coding agents:
- **Instant Single-Binary CLI (< 15MB RAM)**: Installs with `go install` or `brew install`. Launches an ephemeral local HTTP server and opens the browser in under **10 milliseconds**.
- **Universal Multi-Artifact Annotation**:
  - **Markdown Plans & Specs**: Highlight text, click to leave comment balloons, apply redline strikethroughs.
  - **Local Git Diffs**: Side-by-side and unified code reviews with line-level code suggestions (identical to GitHub PR reviews, but 100% local).
  - **HTML & Web Component Previews**: Sandboxed iframe preview with responsive device viewports (Desktop, Tablet, Mobile) and point-and-click UI issue pins.
  - **Interactive Architecture Diagrams**: Live interactive Mermaid and SVG rendering with visual pins on diagram nodes.
- **Surgical Patch Feedback Generator**: Instead of sending vague text, `planmark` packages all annotations into **deterministic, line-anchored surgical patch blocks** that agents parse with 100% precision.
- **Multi-Agent Session Inbox**: A unified workspace organizing parallel agent review requests into tabs so developers can approve or annotate work asynchronously without blocking terminal sessions.

---

## 2. High-Level System Architecture

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
       |                    planmark Core CLI                   |
       |          (Single Go Binary, Sub-10ms Startup)          |
       +--------------------------------------------------------+
                                   |
                     +-------------+-------------+
                     |                           |
                     v                           v
     +-------------------------------+   +-------------------------------+
     |      Ephemeral HTTP Server    |   |     Document AST & Parser     |
     | - Localhost random port/sock  |   | - Markdown (goldmark AST)     |
     | - WebSocket live sync         |   | - Git Unified Diff parser     |
     | - Embedded React 19 assets    |   | - Mermaid & SVG DOM renderer  |
     +-------------------------------+   +-------------------------------+
                     |                           |
                     +-------------+-------------+
                                   |
                                   v
       +--------------------------------------------------------+
       |               React 19 Visual Review Canvas            |
       |  - Mode 1: Document & Plan Annotation (Highlights)     |
       |  - Mode 2: Git Diff Code Review (Line Suggestions)     |
       |  - Mode 3: Interactive HTML / Component Sandbox        |
       |  - Mode 4: Diagram Pinning (Mermaid & Architecture)    |
       +--------------------------------------------------------+
                                   |
                                   | User clicks: [Send Feedback to Agent]
                                   v
       +--------------------------------------------------------+
       |           Surgical Feedback Formatter Engine           |
       |  - Generates Line-Anchored Mutation Blocks             |
       |  - Builds Deterministic Diff Suggestions               |
       |  - Synthesizes Consolidated Agent Prompt               |
       +--------------------------------------------------------+
                                   |
                                   | Exits with formatted output or sends via
                                   | IPC / stdin back to running agent
                                   v
       +--------------------------------------------------------+
       |             Agent Executes Approved Revisions          |
       +--------------------------------------------------------+
```

---

## 3. The Surgical Patch Protocol

Unlike primitive feedback tools that output unstructured paragraphs, `planmark` translates visual annotations into **Surgical Patch Blocks** that LLM agents apply with zero ambiguity:

```markdown
<!-- planmark:feedback-packet v1 -->
# Human Review Feedback on `docs/ARCHITECTURE_PLAN.md`

## Annotation 1: Text Replacement (Lines 42-45)
<<<<<<< ORIGINAL
We will use a central Redis cluster for session synchronization and distributed locks.
======= PROPOSED
We will use SQLite in WAL mode with local in-memory mutexes to eliminate external infrastructure dependencies.
>>>>>>> RATIONALE
Avoid adding Redis operational overhead. Keep deployment single-binary.

## Annotation 2: Code Review Suggestion (internal/auth/token.go:78)
```suggestion
func VerifyToken(ctx context.Context, tokenStr string) (*Claims, error) {
    if tokenStr == "" {
        return nil, ErrEmptyToken
    }
    return parseJWT(ctx, tokenStr)
}
```
*Comment: Add context propagation and return typed error for empty tokens.*

## Annotation 3: Architecture Diagram Pin (Component: [Database Layer])
- Issue: "PostgreSQL node must be marked as optional read-replica, primary storage is SQLite."
<!-- planmark:end-packet -->
```

---

## 4. Universal Agent Hook & Integration

### 4.1 Claude Code Integration (`CLAUDE.md` / Hook)
In project `.claude/config.json` or `CLAUDE.md`:
```bash
# Register planmark as the review hook for Claude Code
planmark hook install claude-code
```
When Claude calls the plan or review step, it executes `planmark review --file plan.md` and pauses until the developer clicks **"Submit Feedback"** in the browser.

### 4.2 Generic Stdin / Pipe Mode (Any CLI Agent)
```bash
# Pipe any agent plan or git diff directly into planmark
cat architecture_spec.md | planmark | claude

# Review uncommitted git changes visually before agent commits
git diff | planmark --diff | aider --stdin
```

### 4.3 Antigravity & IDE Protocol
Exposes local JSON-RPC / WebSocket endpoints so agents running inside Antigravity IDE or Cursor can trigger review modals natively.

---

## 5. Repository & Directory Structure

```
planmark/
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── release.yml
├── cmd/
│   └── planmark/
│       └── main.go                 # CLI entrypoint (review, diff, hook, inbox, serve)
├── internal/
│   ├── server/                     # Local High-Speed Ephemeral Server
│   │   ├── server.go               # HTTP & WebSocket handler
│   │   ├── ports.go                # Dynamic port allocator & unix socket fallback
│   │   └── browser.go              # Cross-platform browser launcher (macOS/Win/Linux)
│   ├── parser/                     # Artifact & Document Parsers
│   │   ├── markdown.go             # goldmark AST line mapper
│   │   ├── diff.go                 # Unified diff & multi-file patch parser
│   │   ├── html_sandbox.go         # CSP-isolated iframe wrapper
│   │   └── mermaid.go              # Mermaid diagram syntax validator
│   ├── format/                     # Surgical Feedback Compiler
│   │   ├── patch_builder.go        # Generates <<<<<<< ======= >>>>>>> blocks
│   │   ├── suggestion.go           # Generates GitHub-style code suggestions
│   │   └── packet.go               # Structured feedback packet assembler
│   ├── hooks/                      # Universal Agent Hook Integrators
│   │   ├── claude.go               # Claude Code plugin & hook registrar
│   │   ├── codex.go                # Codex CLI hook
│   │   ├── opencode.go             # OpenCode integration
│   │   └── aider.go                # Aider pipe handler
│   ├── store/                      # In-Memory & Local Session Cache
│   │   ├── session.go              # Active review sessions
│   │   └── annotations.go          # Live visual pins, highlights, and comments
│   └── embed.go                    # //go:embed assets for React 19 UI
├── web/                            # Embedded React 19 Review Surface
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── App.tsx
│       ├── components/
│       │   ├── PlanAnnotator.tsx   # Markdown visual text highlight & comment cards
│       │   ├── DiffReviewer.tsx    # Split / unified code diff reviewer
│       │   ├── ComponentFrame.tsx  # Responsive HTML mockup sandboxed preview
│       │   ├── DiagramPins.tsx     # Mermaid SVG visual pin annotator
│       │   └── SessionTabs.tsx     # Multi-agent inbox tab switcher
│       └── pages/
│           ├── ReviewWorkspace.tsx
│           └── AgentInbox.tsx
├── examples/
│   ├── sample_plan.md
│   └── sample_diff.patch
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 6. CLI Command Surface

```bash
# Review an architectural plan or markdown spec in browser
planmark review docs/PLAN.md

# Review local uncommitted Git changes visually
planmark diff

# Review specific Git branches or commits
planmark diff origin/main..HEAD

# Launch the unified Multi-Agent Inbox dashboard
planmark inbox

# Install agent hook integrations into current repository
planmark hook install --agent claude-code

# Pipe mode: read from stdin and output approved feedback to stdout
cat draft_plan.md | planmark --pipe
```

---

## 7. Interactive Review Modes in Web Canvas

1. **Plan & Document Mode**:
   - Renders GitHub-flavored Markdown with typography styling.
   - Text selection popover: `[Add Comment]`, `[Suggest Replacement]`, `[Delete / Strike]`.
   - Floating comment sidebar anchored directly to document headings.
2. **Code Diff Mode**:
   - Side-by-side or unified view with syntax highlighting.
   - Click `+` on any line gutter to open an inline code suggestion box.
   - File tree navigation with changed line counters (`+42 / -12`).
3. **HTML / Component Sandbox Mode**:
   - Renders live HTML/Tailwind/CSS in an isolated iframe.
   - Responsive viewport toggles: Mobile (390px), Tablet (768px), Desktop (1440px).
   - "Inspect & Pin" tool: Click on any rendered UI element to add a visual feedback marker.
4. **Diagram & Architecture Mode**:
   - Renders Mermaid sequence, flowchart, and class diagrams into crisp SVG.
   - Allows placing sticky feedback pins directly onto diagram nodes and connections.

---

## 8. Implementation Roadmap (5 Phases)

### Phase 1: Core Daemon & Ephemeral Server
- Implement lightweight Go HTTP server with dynamic port negotiation.
- Implement cross-platform default browser opener (Windows, macOS, Linux).
- Set up stdin/stdout pipe mode and WebSocket protocol for instant UI synchronization.

### Phase 2: Markdown & Spec Annotation Surface
- Implement `goldmark` AST parser mapping every markdown node to source line ranges.
- Build React 19 text selection listener with inline floating toolbar.
- Build surgical patch formatter generating `<<<<<<< ORIGINAL ======= PROPOSED >>>>>>>` blocks.

### Phase 3: Git Diff Code Reviewer
- Implement high-speed unified diff parser in Go (`internal/parser/diff.go`).
- Build split and unified visual diff viewer in React 19 with syntax highlighting.
- Add GitHub-style inline `suggestion` editor with instant patch preview.

### Phase 4: HTML Sandbox & Mermaid Diagram Pinning
- Build CSP-safe iframe component renderer with responsive viewport controls.
- Integrate Mermaid.js SVG renderer with click-to-pin coordinate anchors.
- Add Multi-Agent Inbox navigation allowing asynchronous review of multiple parallel agents.

### Phase 5: Agent Plugins, Packaging & Release
- Build automated hook installers for Claude Code, Gemini CLI, and Codex.
- Embed frontend build into single Go binary using `//go:embed`.
- Verify memory footprint remains strictly under 15MB RAM and startup finishes under 10ms.
- Produce documentation, animated demo GIFs, and release via Homebrew / GitHub Releases.
