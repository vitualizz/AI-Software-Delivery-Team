package setup

import "testing"

// TestNewerAvailable exercises the unexported newerAvailable comparison
// helper directly (white-box) since it must remain unexported per design.
func TestNewerAvailable(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{name: "dev build", current: "dev", latest: "v9.9.9", want: false},
		{name: "empty current", current: "", latest: "v1.2.3", want: false},
		{name: "equal normalized", current: "0.3.0", latest: "v0.3.0", want: false},
		{name: "newer", current: "v0.2.0", latest: "v0.3.0", want: true},
		{name: "empty latest", current: "v0.2.0", latest: "", want: false},
		{name: "whitespace tolerance", current: " v0.2.0 ", latest: "v0.3.0", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newerAvailable(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("newerAvailable(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}
