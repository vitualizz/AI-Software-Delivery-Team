package artifact

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FSStore is the filesystem adapter that implements Store.
// All artifacts are stored under {root}/.asdt/artifacts/{change}/{type}.yaml.
type FSStore struct {
	root string // absolute path to the .asdt/ directory
}

// NewFSStore returns an FSStore rooted at the given .asdt/ directory path.
func NewFSStore(root string) *FSStore {
	return &FSStore{root: root}
}

// artifactPath returns the absolute path for a given change and artifact type.
func (s *FSStore) artifactPath(change, artifactType string) string {
	return filepath.Join(s.root, "artifacts", change, artifactType+".yaml")
}


// Read deserializes the YAML artifact at change/artifactType into out.
func (s *FSStore) Read(_ context.Context, change, artifactType string, out any) error {
	p := s.artifactPath(change, artifactType)
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("artifact read %s/%s: %w", change, artifactType, err)
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("artifact unmarshal %s/%s: %w", change, artifactType, err)
	}
	return nil
}

// Write serializes env to YAML and persists it at change/artifactType.
func (s *FSStore) Write(_ context.Context, change, artifactType string, env any) error {
	p := s.artifactPath(change, artifactType)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("artifact mkdir %s/%s: %w", change, artifactType, err)
	}
	data, err := yaml.Marshal(env)
	if err != nil {
		return fmt.Errorf("artifact marshal %s/%s: %w", change, artifactType, err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return fmt.Errorf("artifact write %s/%s: %w", change, artifactType, err)
	}
	return nil
}

// List returns the artifact types (without .yaml extension) for a given change.
func (s *FSStore) List(_ context.Context, change string) ([]string, error) {
	dir := filepath.Join(s.root, "artifacts", change)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("artifact list %s: %w", change, err)
	}
	var result []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			result = append(result, strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	return result, nil
}

// Exists returns true when the artifact file exists on disk.
func (s *FSStore) Exists(change, artifactType string) bool {
	_, err := os.Stat(s.artifactPath(change, artifactType))
	return err == nil
}
