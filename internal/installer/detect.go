package installer

import (
	"errors"
	"os"
	"os/exec"
)

// Detect reports whether the assistant's binary is in PATH and whether its
// skills directory exists on disk. Neither being false is not an error.
func Detect(d AssistantDescriptor) (binaryPresent bool, skillsPresent bool, err error) {
	_, lookErr := exec.LookPath(d.BinaryName)
	binaryPresent = lookErr == nil

	_, statErr := os.Stat(d.SkillsDir)
	skillsPresent = statErr == nil || !errors.Is(statErr, os.ErrNotExist)
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return binaryPresent, false, statErr
	}
	if statErr != nil {
		skillsPresent = false
	}

	return binaryPresent, skillsPresent, nil
}
