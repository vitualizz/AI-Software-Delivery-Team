package artifact

import (
	"errors"
	"fmt"
	"strings"
)

// ErrSchemaVersionMismatch is returned when an artifact's schema_version
// does not match CurrentSchemaVersion.
var ErrSchemaVersionMismatch = errors.New("schema_version mismatch")

// Validate checks that all required fields of an EnvelopeHeader are present
// and that the schema_version matches CurrentSchemaVersion.
// Returns a descriptive error listing all missing fields, or nil on success.
func Validate(h EnvelopeHeader) error {
	var missing []string

	if h.SchemaVersion == "" {
		missing = append(missing, "schema_version")
	}
	if h.Agent == "" {
		missing = append(missing, "agent")
	}
	if h.ChangeID == "" {
		missing = append(missing, "change_id")
	}
	if h.CreatedAt.IsZero() {
		missing = append(missing, "created_at")
	}
	if h.PromptVersion == "" {
		missing = append(missing, "prompt_version")
	}
	// input_refs may be an empty slice — presence of the field is not validated
	// here because an empty list is valid for the initial artifact in a chain.

	if len(missing) > 0 {
		return fmt.Errorf("envelope missing required fields: %s", strings.Join(missing, ", "))
	}

	if h.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("%w: got %q, want %q",
			ErrSchemaVersionMismatch, h.SchemaVersion, CurrentSchemaVersion)
	}

	return nil
}
