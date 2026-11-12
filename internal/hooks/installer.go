package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type HookInstaller struct{}

func NewHookInstaller() *HookInstaller {
	return &HookInstaller{}
}

func (h *HookInstaller) Install(agent string) error {
	switch agent {
	case "claude-code", "claude":
		return h.installClaudeCode()
	case "gemini", "gemini-cli":
		return h.installGeminiCLI()
	case "codex":
		return h.installCodex()
	case "aider":
		return h.installAider()
	default:
		return fmt.Errorf("unknown agent '%s'. Supported agents: claude-code, gemini-cli, codex, aider", agent)
	}
}

func (h *HookInstaller) installClaudeCode() error {
	claudeDir := ".claude"
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(claudeDir, "config.json")
	var config map[string]any = make(map[string]any)

	if data, err := os.ReadFile(configFile); err == nil {
		_ = json.Unmarshal(data, &config)
	}

	tools, ok := config["tools"].(map[string]any)
	if !ok {
		tools = make(map[string]any)
	}

	tools["review_plan"] = map[string]string{
		"command":     "planmark review --file {file}",
		"description": "Opens the proposed plan in the local PlanMark browser review canvas for human approval.",
	}
	config["tools"] = tools

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func (h *HookInstaller) installGeminiCLI() error {
	geminiDir := ".gemini"
	if err := os.MkdirAll(geminiDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(geminiDir, "tools.json")
	var tools []map[string]any

	if data, err := os.ReadFile(configFile); err == nil {
		_ = json.Unmarshal(data, &tools)
	}

	tools = append(tools, map[string]any{
		"name":        "planmark_review",
		"command":     "planmark review --file $FILE",
		"description": "Interactive visual spec and plan review surface.",
	})

	data, err := json.MarshalIndent(tools, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func (h *HookInstaller) installCodex() error {
	cfgFile := ".codexrc"
	content := "\n# PlanMark review hook\nplan_review_cmd=\"planmark review --file $FILE\"\n"
	f, err := os.OpenFile(cfgFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func (h *HookInstaller) installAider() error {
	cfgFile := ".aider.conf.yml"
	content := "\n# PlanMark review hook\nlint-cmd: \"planmark diff\"\n"
	f, err := os.OpenFile(cfgFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
