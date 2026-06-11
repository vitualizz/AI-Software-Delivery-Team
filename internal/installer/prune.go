package installer

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// pruneStale removes files under the assistant's managed sibling roots that
// the current install no longer provides, returning the SkillsDir-relative
// paths actually removed. Candidate selection prefers the previous install
// manifest (prevFiles) when present; legacy installs without a manifest fall
// back to scanning the managed roots on disk. Every candidate must pass the
// lexical safety gate (underManagedRoot) and the dotfile guard before
// removal. Removal is best-effort: failures skip the candidate and never
// surface as an install error.
func pruneStale(skillsDir string, managedRoots []string, prevFiles, currentFiles []string) []string {
	if len(managedRoots) == 0 {
		return nil
	}

	current := make(map[string]bool, len(currentFiles))
	for _, f := range currentFiles {
		current[f] = true
	}

	var candidates []string
	if len(prevFiles) > 0 {
		for _, f := range prevFiles {
			if !current[f] {
				candidates = append(candidates, f)
			}
		}
	} else {
		candidates = collectFallbackCandidates(skillsDir, managedRoots, current)
	}

	var removed []string
	for _, rel := range candidates {
		if hasDotComponent(rel) || !underManagedRoot(rel, managedRoots) {
			continue
		}
		if err := os.Remove(filepath.Join(skillsDir, rel)); err != nil {
			continue
		}
		removed = append(removed, rel)
	}

	removeEmptyParents(skillsDir, removed, managedRoots)
	return removed
}

// managedRootsFor derives the deduplicated set of sibling directory names
// under an assistant's SkillsDir that ASDT owns, from the top-level entries
// of the embedded skill tree. Directory entries map verbatim; any loose
// root-level file contributes the consultant's "asdt" root. A ReadDir error
// yields nil, which disables pruning entirely.
func managedRootsFor(skillsFS fs.FS) []string {
	entries, err := fs.ReadDir(skillsFS, ".")
	if err != nil {
		return nil
	}

	seen := make(map[string]bool, len(entries))
	var roots []string
	for _, entry := range entries {
		name := SiblingDestName(".")
		if entry.IsDir() {
			name = SiblingDestName(entry.Name())
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		roots = append(roots, name)
	}
	return roots
}

// relativizeWritten converts the absolute paths recorded in
// InstallResult.Written into SkillsDir-relative paths. Entries that cannot
// be relativized or that escape skillsDir are dropped.
func relativizeWritten(skillsDir string, written []string) []string {
	var rels []string
	for _, abs := range written {
		rel, err := filepath.Rel(skillsDir, abs)
		if err != nil {
			continue
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		rels = append(rels, rel)
	}
	return rels
}

// collectFallbackCandidates scans each managed root that exists on disk and
// collects the SkillsDir-relative paths of files (and symlinks) that the
// current install did not write and that are not dotfile-protected. It never
// visits siblings outside the managed roots.
func collectFallbackCandidates(skillsDir string, managedRoots []string, current map[string]bool) []string {
	var candidates []string
	for _, root := range managedRoots {
		rootAbs := filepath.Join(skillsDir, root)
		if _, err := os.Lstat(rootAbs); err != nil {
			continue
		}
		_ = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // best-effort: skip unreadable entries
			}
			if d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(skillsDir, path)
			if relErr != nil {
				return nil
			}
			if current[rel] || hasDotComponent(rel) {
				return nil
			}
			candidates = append(candidates, rel)
			return nil
		})
	}
	return candidates
}

// underManagedRoot reports whether rel is a relative path strictly inside
// one of the managed roots: not absolute, no ".." components, and at least
// one component below a managed root (the root itself never qualifies).
func underManagedRoot(rel string, managedRoots []string) bool {
	if rel == "" || filepath.IsAbs(rel) {
		return false
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, part := range parts {
		if part == ".." {
			return false
		}
	}
	if len(parts) < 2 {
		return false
	}
	for _, root := range managedRoots {
		if parts[0] == root {
			return true
		}
	}
	return false
}

// hasDotComponent reports whether ANY path component of rel starts with a
// dot — protecting metadata such as .install-meta.json at every depth.
func hasDotComponent(rel string) bool {
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

// removeEmptyParents ascends each removed file's parent chain strictly below
// its managed root, removing directories left empty by the prune. The first
// failure on a chain (typically ENOTEMPTY) stops that chain. Managed roots
// themselves and skillsDir are never removed.
func removeEmptyParents(skillsDir string, removedRel []string, managedRoots []string) {
	for _, rel := range removedRel {
		for dir := filepath.Dir(rel); underManagedRoot(dir, managedRoots); dir = filepath.Dir(dir) {
			if os.Remove(filepath.Join(skillsDir, dir)) != nil {
				break
			}
		}
	}
}
