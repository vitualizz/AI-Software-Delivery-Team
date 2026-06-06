package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/config"
)

func makeCfgRoot(t *testing.T) config.Root {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".asdt"), 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("config.Discover: %v", err)
	}
	return root
}
