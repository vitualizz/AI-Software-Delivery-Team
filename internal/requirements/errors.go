package requirements

import (
	"errors"
	"fmt"
)

// ErrMissingIdea is returned when Run is called with an empty idea string.
var ErrMissingIdea = errors.New("requirements: idea is required")

// ErrAmbiguousIdea is returned when the idea is too vague to produce a spec
// (fewer than 5 words). Question carries a clarifying question for the caller
// to surface to the user.
type ErrAmbiguousIdea struct {
	Question string
}

func (e ErrAmbiguousIdea) Error() string {
	return fmt.Sprintf("requirements: ambiguous idea — %s", e.Question)
}
