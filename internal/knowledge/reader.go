package knowledge

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vitualizz/asdt/internal/config"
	"gopkg.in/yaml.v3"
)

// Reader is the port for reading and writing platform.yaml and project-context.yaml.
// Both files live at {root}/knowledge/ — outside the per-change artifact tree
// because they describe the whole project.
type Reader interface {
	// Read loads platform.yaml from the given .asdt/ root.
	// Returns an error if the file does not exist or cannot be parsed.
	Read(root config.Root) (Platform, error)

	// Write serializes p and persists it at {root}/knowledge/platform.yaml.
	Write(root config.Root, p Platform) error

	// WriteSummary serializes s and persists it at
	// {root}/knowledge/platform-summary.yaml. Creates the directory if needed.
	WriteSummary(root config.Root, s PlatformSummary) error

	// WriteContext serializes ctx and persists it at
	// {root}/knowledge/project-context.yaml. Creates the directory if needed.
	WriteContext(root config.Root, ctx ProjectContext) error

	// ReadContext loads project-context.yaml from the given .asdt/ root.
	// Returns an error if the file does not exist or cannot be parsed.
	ReadContext(root config.Root) (ProjectContext, error)
}

// FSReader is the filesystem-backed implementation of Reader.
type FSReader struct{}

// NewFSReader constructs an FSReader. The store parameter is accepted for
// interface compatibility but FSReader manages its own path under {root}/knowledge/.
func NewFSReader() *FSReader {
	return &FSReader{}
}

// platformPath returns the absolute path to platform.yaml for the given root.
func platformPath(root config.Root) string {
	return filepath.Join(root.Path(), "knowledge", "platform.yaml")
}

// Read loads and deserializes platform.yaml from {root}/knowledge/platform.yaml.
func (r *FSReader) Read(root config.Root) (Platform, error) {
	p := platformPath(root)
	data, err := os.ReadFile(p)
	if err != nil {
		return Platform{}, fmt.Errorf("knowledge read: %w", err)
	}
	var platform Platform
	if err := yaml.Unmarshal(data, &platform); err != nil {
		return Platform{}, fmt.Errorf("knowledge unmarshal: %w", err)
	}
	return platform, nil
}

// Write serializes p to YAML and writes it to {root}/knowledge/platform.yaml.
// Creates the directory if it does not exist.
func (r *FSReader) Write(root config.Root, p Platform) error {
	path := platformPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("knowledge mkdir: %w", err)
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("knowledge marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("knowledge write: %w", err)
	}
	return nil
}

// summaryPath returns the absolute path to platform-summary.yaml for the given root.
func summaryPath(root config.Root) string {
	return filepath.Join(root.Path(), "knowledge", "platform-summary.yaml")
}

// contextPath returns the absolute path to project-context.yaml for the given root.
func contextPath(root config.Root) string {
	return filepath.Join(root.Path(), "knowledge", "project-context.yaml")
}

// WriteSummary serializes s to YAML and writes it to
// {root}/knowledge/platform-summary.yaml. Creates the directory if needed.
func (r *FSReader) WriteSummary(root config.Root, s PlatformSummary) error {
	path := summaryPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("knowledge summary mkdir: %w", err)
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("knowledge summary marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("knowledge summary write: %w", err)
	}
	return nil
}

// WriteContext serializes ctx to YAML and writes it to
// {root}/knowledge/project-context.yaml. Creates the directory if needed.
func (r *FSReader) WriteContext(root config.Root, ctx ProjectContext) error {
	path := contextPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("knowledge context mkdir: %w", err)
	}
	data, err := yaml.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("knowledge context marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("knowledge context write: %w", err)
	}
	return nil
}

// ReadContext loads and deserializes project-context.yaml from
// {root}/knowledge/project-context.yaml. Returns an error if the file is absent.
func (r *FSReader) ReadContext(root config.Root) (ProjectContext, error) {
	p := contextPath(root)
	data, err := os.ReadFile(p)
	if err != nil {
		return ProjectContext{}, fmt.Errorf("knowledge context read: %w", err)
	}
	var ctx ProjectContext
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return ProjectContext{}, fmt.Errorf("knowledge context unmarshal: %w", err)
	}
	return ctx, nil
}
