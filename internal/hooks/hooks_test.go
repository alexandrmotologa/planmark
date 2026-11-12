package hooks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/planmark/internal/hooks"
)

func TestHookInstaller(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "planmark_hook_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(origWd)

	installer := hooks.NewHookInstaller()

	// Test installing claude-code
	if err := installer.Install("claude-code"); err != nil {
		t.Fatalf("failed to install claude-code hook: %v", err)
	}

	claudeConfig := filepath.Join(".claude", "config.json")
	if _, err := os.Stat(claudeConfig); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", claudeConfig)
	}

	// Test installing gemini-cli
	if err := installer.Install("gemini-cli"); err != nil {
		t.Fatalf("failed to install gemini-cli hook: %v", err)
	}

	geminiConfig := filepath.Join(".gemini", "tools.json")
	if _, err := os.Stat(geminiConfig); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", geminiConfig)
	}

	// Test installing aider
	if err := installer.Install("aider"); err != nil {
		t.Fatalf("failed to install aider hook: %v", err)
	}

	aiderConfig := ".aider.conf.yml"
	if _, err := os.Stat(aiderConfig); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", aiderConfig)
	}

	// Test unknown agent returns error
	if err := installer.Install("unsupported-agent"); err == nil {
		t.Errorf("expected error for unsupported agent")
	}
}
