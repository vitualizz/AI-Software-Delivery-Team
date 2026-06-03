package developer

import "errors"

// ErrMissingRequirements is returned by Run when no requirements-spec artifact
// exists for the given change. The caller should instruct the user to run
// the requirements agent first.
var ErrMissingRequirements = errors.New("developer: requirements-spec not found; run the requirements agent first")
