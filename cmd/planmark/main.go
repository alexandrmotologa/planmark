package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/planmark"
	"github.com/alexandrmotologa/planmark/internal/hooks"
	"github.com/alexandrmotologa/planmark/internal/mcp"
	"github.com/alexandrmotologa/planmark/internal/parser"
	"github.com/alexandrmotologa/planmark/internal/server"
	"github.com/alexandrmotologa/planmark/internal/store"
	"github.com/alexandrmotologa/planmark/internal/types"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		// Check if data is being piped into stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			runPipeMode("", false)
			return
		}
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "review":
		runReview(args)
	case "diff":
		runDiff(args)
	case "inbox":
		runInbox(args)
	case "hook":
		runHook(args)
	case "mcp":
		runMCP()
	case "version", "--version", "-v":
		fmt.Printf("planmark version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		// If first argument is a file that exists, treat as "planmark review <file>"
		if _, err := os.Stat(command); err == nil {
			runReview(os.Args[1:])
			return
		}
		fmt.Fprintf(os.Stderr, "Unknown command '%s'. Run 'planmark help' for usage.\n", command)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`planmark — Local-First Visual Review & Surgical Patch Feedback Surface

Usage:
  planmark review <file> [flags]   Review a spec, plan, HTML component or diagram
  planmark diff [git-args...]      Review local uncommitted changes or branch diffs
  planmark inbox [flags]           Launch persistent multi-agent review inbox
  planmark hook install [flags]    Configure review hooks for AI coding agents
  planmark mcp                     Run as Model Context Protocol stdio server
  cat plan.md | planmark           Pipe mode for Unix agent pipelines

Flags:
  --port <port>     Override ephemeral server port (default: random open port)
  --json            Output review feedback as JSON instead of surgical markdown
  --no-browser      Do not automatically open default web browser
  --dev             Development mode: redirect to Vite dev server (http://localhost:5173)`)
}

func runReview(args []string) {
	fs := flag.NewFlagSet("review", flag.ExitOnError)
	portFlag := fs.Int("port", 0, "Server port")
	jsonFlag := fs.Bool("json", false, "Output JSON feedback")
	noBrowserFlag := fs.Bool("no-browser", false, "Do not launch browser")
	devFlag := fs.Bool("dev", false, "Run with Vite dev server")
	fileFlag := fs.String("file", "", "Target file to review")

	_ = fs.Parse(args)

	filePath := *fileFlag
	if filePath == "" && len(fs.Args()) > 0 {
		filePath = fs.Args()[0]
	}

	if filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: No file specified. Usage: planmark review <file>")
		os.Exit(1)
	}

	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", filePath, err)
		os.Exit(1)
	}
	content := string(contentBytes)

	artType := detectArtifactType(filePath, content)
	session := &types.Session{
		Agent: "manual",
		Title: filepath.Base(filePath),
		Artifact: types.Artifact{
			Type:     artType,
			Title:    filepath.Base(filePath),
			FilePath: filePath,
			Content:  content,
		},
	}

	enrichSessionArtifact(session)
	runSessionInteractive(session, *portFlag, *jsonFlag, *noBrowserFlag, *devFlag)
}

func runDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	portFlag := fs.Int("port", 0, "Server port")
	jsonFlag := fs.Bool("json", false, "Output JSON feedback")
	noBrowserFlag := fs.Bool("no-browser", false, "Do not launch browser")
	devFlag := fs.Bool("dev", false, "Run with Vite dev server")

	_ = fs.Parse(args)
	gitArgs := fs.Args()

	cmdArgs := []string{"diff"}
	if len(gitArgs) > 0 {
		cmdArgs = append(cmdArgs, gitArgs...)
	}

	cmd := exec.Command("git", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing 'git diff': %v\n%s\n", err, stderr.String())
		os.Exit(1)
	}

	diffOutput := stdout.String()
	if strings.TrimSpace(diffOutput) == "" {
		fmt.Println("No git diff detected. Working tree clean.")
		return
	}

	session := &types.Session{
		Agent: "git-review",
		Title: "Git Diff Review",
		Artifact: types.Artifact{
			Type:     types.ArtifactDiff,
			Title:    "Git Diff",
			FilePath: "git.patch",
			Content:  diffOutput,
		},
	}

	enrichSessionArtifact(session)
	runSessionInteractive(session, *portFlag, *jsonFlag, *noBrowserFlag, *devFlag)
}

func runPipeMode(forcedType types.ArtifactType, outputJSON bool) {
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
	content := string(inputBytes)
	if strings.TrimSpace(content) == "" {
		fmt.Fprintln(os.Stderr, "Error: stdin was empty.")
		os.Exit(1)
	}

	artType := forcedType
	if artType == "" {
		artType = detectArtifactType("stdin", content)
	}

	session := &types.Session{
		Agent: "pipeline",
		Title: "Piped Artifact Review",
		Artifact: types.Artifact{
			Type:     artType,
			Title:    "Piped Artifact",
			FilePath: "stdin",
			Content:  content,
		},
	}

	enrichSessionArtifact(session)
	runSessionInteractive(session, 0, outputJSON, false, false)
}

func runInbox(args []string) {
	fs := flag.NewFlagSet("inbox", flag.ExitOnError)
	portFlag := fs.Int("port", 8765, "Server port")
	noBrowserFlag := fs.Bool("no-browser", false, "Do not launch browser")
	_ = fs.Parse(args)

	st := store.NewSessionStore()
	webAssets := planmark.GetSubFS()

	srv := server.NewServer(server.ServerConfig{
		Port:         *portFlag,
		Store:        st,
		Assets:       webAssets,
		AutoShutdown: false,
	})

	url := fmt.Sprintf("http://127.0.0.1:%d/inbox", *portFlag)
	fmt.Printf("PlanMark Multi-Agent Inbox running on %s\n", url)
	if !*noBrowserFlag {
		_ = server.OpenBrowser(url)
	}

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}

func runHook(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: planmark hook install [--agent claude-code|gemini-cli|codex|aider]")
		os.Exit(1)
	}

	action := args[0]
	if action != "install" {
		fmt.Fprintf(os.Stderr, "Unknown hook action '%s'\n", action)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("hook install", flag.ExitOnError)
	agentFlag := fs.String("agent", "claude-code", "Target AI agent")
	_ = fs.Parse(args[1:])

	installer := hooks.NewHookInstaller()
	if err := installer.Install(*agentFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error installing hook for '%s': %v\n", *agentFlag, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully registered PlanMark hook for %s.\n", *agentFlag)
}

func runMCP() {
	st := store.NewSessionStore()
	webAssets := planmark.GetSubFS()
	mcpServer := mcp.NewMCPServer(st, webAssets)
	if err := mcpServer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP Server error: %v\n", err)
		os.Exit(1)
	}
}

func runSessionInteractive(session *types.Session, port int, outputJSON bool, noBrowser bool, devMode bool) {
	st := store.NewSessionStore()
	created := st.Create(session)

	if port == 0 {
		var err error
		port, err = server.GetFreePort()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error allocating port: %v\n", err)
			os.Exit(1)
		}
	}

	webAssets := planmark.GetSubFS()
	doneChan := make(chan *types.FeedbackPacket, 1)

	srv := server.NewServer(server.ServerConfig{
		Port:         port,
		Store:        st,
		Assets:       webAssets,
		DoneChan:     doneChan,
		AutoShutdown: true,
	})

	go func() {
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	var appURL string
	if devMode {
		appURL = fmt.Sprintf("http://localhost:5173/session/%s?backend=%d", created.ID, port)
	} else {
		appURL = fmt.Sprintf("http://127.0.0.1:%d/session/%s", port, created.ID)
	}

	fmt.Fprintf(os.Stderr, "PlanMark studio opened at %s\nWaiting for review submission...\n", appURL)
	if !noBrowser {
		_ = server.OpenBrowser(appURL)
	}

	// Block until user submits feedback
	packet := <-doneChan
	_ = srv.Shutdown()

	if outputJSON {
		// Output structured JSON
		fmt.Println(packet.RawMarkdown) // or JSON
	} else {
		// Output canonical surgical patch packet
		fmt.Print(packet.RawMarkdown)
	}
}

func detectArtifactType(filePath, content string) types.ArtifactType {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".patch", ".diff":
		return types.ArtifactDiff
	case ".html", ".htm":
		return types.ArtifactHTML
	case ".mmd", ".mermaid":
		return types.ArtifactMermaid
	case ".md", ".markdown":
		return types.ArtifactMarkdown
	}

	firstLine := strings.TrimSpace(strings.Split(content, "\n")[0])
	if strings.HasPrefix(firstLine, "diff --git") || strings.HasPrefix(firstLine, "--- ") {
		return types.ArtifactDiff
	}
	if strings.HasPrefix(strings.ToLower(firstLine), "<!doctype html") || strings.HasPrefix(strings.ToLower(firstLine), "<html") {
		return types.ArtifactHTML
	}
	if strings.HasPrefix(strings.ToLower(firstLine), "sequencediagram") || strings.HasPrefix(strings.ToLower(firstLine), "flowchart") || strings.HasPrefix(strings.ToLower(firstLine), "graph") {
		return types.ArtifactMermaid
	}

	return types.ArtifactMarkdown
}

func enrichSessionArtifact(s *types.Session) {
	switch s.Artifact.Type {
	case types.ArtifactMarkdown:
		if p, err := parser.NewMarkdownParser().Parse(s.Artifact.Content); err == nil {
			s.Artifact.Parsed = p
		}
	case types.ArtifactDiff:
		if p, err := parser.NewDiffParser().Parse(s.Artifact.Content); err == nil {
			s.Artifact.Parsed = p
		}
	case types.ArtifactMermaid:
		if p, err := parser.NewMermaidParser().Parse(s.Artifact.Content); err == nil {
			s.Artifact.Parsed = p
		}
	}
}
