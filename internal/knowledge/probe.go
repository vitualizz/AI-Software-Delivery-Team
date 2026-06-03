package knowledge

import (
	"os"
	"path/filepath"
)

// StackProbe examines a project root and reports whether a particular
// technology stack is present.
type StackProbe interface {
	// Name returns the stack identifier written to detected_stack[].
	Name() string

	// Detect returns true when the stack marker is found under projectRoot.
	Detect(projectRoot string) (bool, error)
}

// fileProbe is a StackProbe that checks for the presence of a single marker file.
type fileProbe struct {
	name    string // stack name (e.g. "go", "node")
	markers []string // filenames to check (any one match → detected)
}

// Name returns the stack identifier.
func (p *fileProbe) Name() string { return p.name }

// Detect returns true when any marker file exists directly under projectRoot.
func (p *fileProbe) Detect(projectRoot string) (bool, error) {
	for _, marker := range p.markers {
		path := filepath.Join(projectRoot, marker)
		_, err := os.Stat(path)
		if err == nil {
			return true, nil
		}
		if !os.IsNotExist(err) {
			return false, err
		}
	}
	return false, nil
}

// GoProbe detects Go projects via go.mod.
func GoProbe() StackProbe {
	return &fileProbe{name: "go", markers: []string{"go.mod"}}
}

// NodeProbe detects Node.js projects via package.json.
func NodeProbe() StackProbe {
	return &fileProbe{name: "node", markers: []string{"package.json"}}
}

// RustProbe detects Rust projects via Cargo.toml.
func RustProbe() StackProbe {
	return &fileProbe{name: "rust", markers: []string{"Cargo.toml"}}
}

// PythonProbe detects Python projects via pyproject.toml or requirements.txt.
func PythonProbe() StackProbe {
	return &fileProbe{name: "python", markers: []string{"pyproject.toml", "requirements.txt"}}
}

// RubyProbe detects Ruby projects via Gemfile.
func RubyProbe() StackProbe {
	return &fileProbe{name: "ruby", markers: []string{"Gemfile"}}
}
