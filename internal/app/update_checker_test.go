package app

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "plain", input: "1.2.3", expected: "1.2.3"},
		{name: "prefixed", input: "v1.2.3", expected: "1.2.3"},
		{name: "suffix", input: "v1.2.3-beta.1", expected: "1.2.3"},
		{name: "spaces", input: "  v2.0.1  ", expected: "2.0.1"},
		{name: "invalid", input: "dev", expected: ""},
	}

	for _, tt := range tests {
		if got := normalizeVersion(tt.input); got != tt.expected {
			t.Fatalf("%s: normalizeVersion(%q) = %q, want %q", tt.name, tt.input, got, tt.expected)
		}
	}
}

func TestIsVersionNewer(t *testing.T) {
	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
	}{
		{name: "new patch", latest: "1.2.4", current: "1.2.3", expected: true},
		{name: "new minor", latest: "1.3.0", current: "1.2.9", expected: true},
		{name: "same", latest: "1.2.3", current: "1.2.3", expected: false},
		{name: "older", latest: "1.2.2", current: "1.2.3", expected: false},
		{name: "different length", latest: "1.2.0", current: "1.2", expected: false},
		{name: "extra component", latest: "1.2.1.1", current: "1.2.1", expected: true},
	}

	for _, tt := range tests {
		if got := isVersionNewer(tt.latest, tt.current); got != tt.expected {
			t.Fatalf("%s: isVersionNewer(%q, %q) = %v, want %v", tt.name, tt.latest, tt.current, got, tt.expected)
		}
	}
}
