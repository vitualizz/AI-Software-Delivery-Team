package installer

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// Detect reports whether the assistant's binary is in PATH and whether ASDT
// is already installed for it. Neither being false is not an error.
//
// IMPORTANT — SkillsDir semantic shift (sibling-install layout):
// AssistantDescriptor.SkillsDir is the assistant's shared skills ROOT (e.g.
// ~/.claude/skills), not an ASDT-specific directory — see assistants.go. That
// root commonly exists for any user of the assistant regardless of whether
// ASDT itself was ever installed (other tools/skills live there too).
//
// Therefore skillsPresent does NOT check os.Stat(d.SkillsDir) directly —
// doing so would only prove "the assistant has a skills directory at all",
// which is nearly always true and would misreport ASDT as installed for any
// user who has unrelated skills. Instead, skillsPresent checks for ASDT's
// OWN consultant directory at {SkillsDir}/asdt — the one fixed, predictable
// marker of "ASDT is installed" under the generic sibling-mapping scheme
// (see installer.SiblingDestName / installOne).
func Detect(d AssistantDescriptor) (binaryPresent bool, skillsPresent bool, err error) {
	_, lookErr := exec.LookPath(d.BinaryName)
	binaryPresent = lookErr == nil

	asdtDir := filepath.Join(d.SkillsDir, siblingConsultantDir)
	_, statErr := os.Stat(asdtDir)
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return binaryPresent, false, statErr
	}
	skillsPresent = statErr == nil

	return binaryPresent, skillsPresent, nil
}
