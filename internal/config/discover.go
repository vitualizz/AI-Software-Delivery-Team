// Package config implements .asdt/ root walk-up discovery and config.yaml
// read/write operations. It does not depend on any other internal package.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// ErrNotFound is returned by Discover when no .asdt/ directory is found
// in any ancestor of startDir.
var ErrNotFound = errors.New("no .asdt/ directory found in directory tree; " +
	"run `/asdt` from your project root or run `mkdir .asdt` to initialize")

// Root represents the absolute path to a discovered .asdt/ directory.
type Root struct {
	path string
}

// Path returns the absolute filesystem path to the .asdt/ directory.
func (r Root) Path() string { return r.path }

// Discover walks up the directory tree starting at startDir, searching
// for a directory named ".asdt/". It returns the first (nearest) ancestor
// that contains ".asdt/", wrapping it as a Root.
//
// Walk-up rules:
//   - Stops at the filesystem root.
//   - MUST NOT cross filesystem mount boundaries (device ID changes).
//   - Returns ErrNotFound if no ancestor contains .asdt/.
func Discover(startDir string) (Root, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return Root{}, fmt.Errorf("config discover: resolve %q: %w", startDir, err)
	}

	// Record the device of the starting directory to detect mount boundaries.
	startInfo, err := os.Stat(abs)
	if err != nil {
		return Root{}, fmt.Errorf("config discover: stat %q: %w", abs, err)
	}
	startDev := deviceID(startInfo)

	current := abs
	for {
		candidate := filepath.Join(current, ".asdt")
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return Root{path: candidate}, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root.
			return Root{}, ErrNotFound
		}

		// Check device boundary before ascending.
		parentInfo, err := os.Stat(parent)
		if err != nil {
			return Root{}, fmt.Errorf("config discover: stat parent %q: %w", parent, err)
		}
		if deviceID(parentInfo) != startDev {
			// Mount boundary crossed — stop.
			return Root{}, ErrNotFound
		}

		current = parent
	}
}

// deviceID extracts the device identifier from an os.FileInfo using syscall.Stat_t.
// Returns 0 on platforms where Sys() does not return *syscall.Stat_t.
func deviceID(info os.FileInfo) uint64 {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		return uint64(sys.Dev)
	}
	return 0
}
