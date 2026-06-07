// Package setup implements the Bubbletea-based installer TUI for ASDT.
package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

// InstallDoneMsg is sent by InstallCmd when the installer goroutine completes.
type InstallDoneMsg struct {
	Results []installer.InstallResult
}

// EngramCheckMsg is sent by EngramCheckCmd with the result of the PATH lookup.
type EngramCheckMsg struct {
	Found bool
}

// EngramCheckCmd returns a tea.Cmd that checks whether the engram binary is on
// PATH and sends EngramCheckMsg with the result.
func EngramCheckCmd() tea.Cmd {
	return func() tea.Msg {
		_, err := exec.LookPath("engram")
		return EngramCheckMsg{Found: err == nil}
	}
}

// UpdateCheckMsg carries the result of the GitHub latest-release lookup.
type UpdateCheckMsg struct {
	Latest  string // tag_name from GitHub, e.g. "v0.3.0" (empty on error)
	Current string // version injected at build time, echoed back for comparison
	Err     error  // non-nil on any network/decode/non-200; banner suppressed
}

const (
	latestReleaseURL   = "https://api.github.com/repos/vitualizz/ai-software-delivery-team/releases/latest"
	updateCheckTimeout = 3 * time.Second
)

// UpdateCheckCmd returns a tea.Cmd that fetches the latest GitHub release tag
// and reports it via UpdateCheckMsg. All errors are carried in the Msg; it
// never panics and never blocks beyond updateCheckTimeout.
func UpdateCheckCmd(current string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
		if err != nil {
			return UpdateCheckMsg{Current: current, Err: err}
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return UpdateCheckMsg{Current: current, Err: err}
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return UpdateCheckMsg{Current: current, Err: fmt.Errorf("github status %d", resp.StatusCode)}
		}
		var payload struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return UpdateCheckMsg{Current: current, Err: err}
		}
		return UpdateCheckMsg{Latest: payload.TagName, Current: current}
	}
}

// newerAvailable reports whether latest is a different, non-empty release than
// current. Fails closed on dev/empty builds: never returns true for an
// unstamped binary. v1 uses normalized inequality, not semver ordering.
func newerAvailable(current, latest string) bool {
	if current == "dev" || current == "" {
		return false // MANDATORY dev-build guard, before any comparison
	}
	c := strings.TrimPrefix(strings.TrimSpace(current), "v")
	l := strings.TrimPrefix(strings.TrimSpace(latest), "v")
	if l == "" {
		return false // fail closed: no banner on empty/malformed remote
	}
	return l != c
}

// InstallCmd returns a tea.Cmd that runs installer.Install asynchronously
// and wraps the results in InstallDoneMsg.
func InstallCmd(assistants []installer.AssistantDescriptor, provider installer.ProviderDescriptor, skillsFS fs.FS) tea.Cmd {
	return func() tea.Msg {
		results := installer.Install(assistants, provider, skillsFS)
		return InstallDoneMsg{Results: results}
	}
}
