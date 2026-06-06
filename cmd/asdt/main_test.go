package main

import (
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/config"
)

func TestBuildMemoryProvider_ReturnsError_WhenProviderEmpty(t *testing.T) {
	cfg := config.Config{} // Memory.Provider == ""
	_, err := buildMemoryProvider(cfg)
	if err == nil {
		t.Fatal("expected error when memory.provider is empty, got nil")
	}
}

func TestBuildMemoryProvider_ReturnsProvider_WhenEngram(t *testing.T) {
	cfg := config.Config{
		Memory: config.MemoryConfig{
			Provider: "engram",
			Project:  "test-project",
		},
	}
	provider, err := buildMemoryProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider, got nil")
	}
	if provider.Name() != "engram" {
		t.Errorf("provider.Name() = %q, want %q", provider.Name(), "engram")
	}
}

func TestBuildMemoryProvider_ReturnsError_WhenUnknownProvider(t *testing.T) {
	cfg := config.Config{
		Memory: config.MemoryConfig{
			Provider: "unknown-backend",
		},
	}
	_, err := buildMemoryProvider(cfg)
	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}
}
