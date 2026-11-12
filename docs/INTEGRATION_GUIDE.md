# Agent Integration Guide

This guide covers connecting PlanMark to AI coding agents, terminal workflows, and IDE extensions.

## Supported Agents

- Claude Code
- Gemini CLI
- OpenAI Codex CLI
- Aider
- OpenCode
- Antigravity IDE & Cursor

---

## 1. Claude Code Integration

### Automatic Hook Installation

Run the following command inside your project root:

```bash
planmark hook install --agent claude-code
```

This creates or updates `.claude/config.json` with a review tool definition that routes plan verification steps through `planmark review`.

### Manual Configuration

Add this tool configuration to your project `.claude/config.json`:

```json
{
  "tools": {
    "review_plan": {
      "command": "planmark review --file {file}",
      "description": "Opens the proposed plan in the local PlanMark browser review canvas for human approval."
    }
  }
}
```

---

## 2. Pipe Workflows (UNIX Stdin / Stdout)

PlanMark can read artifacts directly from standard input and emit approved feedback packets to standard output.

### Reviewing Draft Plans

```bash
cat draft_spec.md | planmark --pipe | claude
```

PlanMark opens the browser, loads the spec, and waits. When you click "Submit Feedback", PlanMark writes the surgical patch packet to standard output and closes.

### Reviewing Uncommitted Git Diffs

```bash
git diff | planmark --diff --pipe | aider --stdin
```

---

## 3. Model Context Protocol (MCP)

PlanMark runs as a standard Model Context Protocol server over stdio.

### Antigravity / Claude Desktop / Cursor Setup

Add the following to your MCP configuration file (`mcp_config.json` or `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "planmark": {
      "command": "planmark",
      "args": ["mcp"]
    }
  }
}
```

### Available MCP Tools

- `request_human_review`: Submits a plan, diff, or HTML markup string for review. The tool call blocks until the developer approves, requests changes, or rejects the review.
- `get_active_sessions`: Lists all pending and completed review sessions in the local inbox.
- `cancel_review`: Dismisses an active review request.

---

## 4. Multi-Agent Inbox Daemon

When working with parallel subagents, launch PlanMark in inbox mode:

```bash
planmark inbox --port 8765
```

Agents send artifacts via HTTP POST to `http://127.0.0.1:8765/api/sessions`:

```bash
curl -X POST http://127.0.0.1:8765/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "security-auditor",
    "title": "Auth Refactor Plan",
    "type": "markdown",
    "content": "# Auth Refactor\n..."
  }'
```

The developer switches between agent tabs in the browser, leaving notes and approving tasks independently.
