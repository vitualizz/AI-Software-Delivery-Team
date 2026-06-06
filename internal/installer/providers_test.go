package installer_test

import (
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

func TestProviders_NonEmpty(t *testing.T) {
	if len(installer.Providers) == 0 {
		t.Fatal("Providers slice is empty")
	}
}

func TestProviders_EngramEntry(t *testing.T) {
	var found bool
	for _, p := range installer.Providers {
		if p.ID == installer.ProviderEngram {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no provider with ID = ProviderEngram found in Providers")
	}
}

func TestProviders_CustomizeSkillIdentity(t *testing.T) {
	for _, p := range installer.Providers {
		if p.ID == installer.ProviderEngram {
			got := p.CustomizeSkill("hello")
			if got != "hello" {
				t.Errorf("engram CustomizeSkill(%q) = %q, want %q", "hello", got, "hello")
			}
			return
		}
	}
	t.Fatal("engram provider not found")
}
