package pkg

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_unmarshallCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{
			name:     "single arg",
			expected: []string{"--arg"},
			value:    "WyItLWFyZyJd",
		},
		{
			name:  "invalid",
			value: "blabla",
		},
		{
			name: "list args",
			expected: []string{
				"/usr/lib/google-golang/bin/go",
				"build", "-mod=vendor", "-trimpath",
				"-tags=netgo",
				"-ldflags=-X main.gitVersion=v1.2.3 -X main.gitSomething=somthg",
			},
			value: "WyIvdXNyL2xpYi9nb29nbGUtZ29sYW5nL2Jpbi9nbyIsImJ1aWxkIiwiLW1vZD12ZW5kb3IiLCItdHJpbXBhdGgiLCItdGFncz1uZXRnbyIsIi1sZGZsYWdzPS1YIG1haW4uZ2l0VmVyc2lvbj12MS4yLjMgLVggbWFpbi5naXRTb21ldGhpbmc9c29tdGhnIl0=",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := unmarshallList(tt.value)
			if err != nil && len(tt.expected) != 0 {
				t.Errorf("marshallList: %v", err)
			}

			if !cmp.Equal(r, tt.expected) {
				t.Errorf(cmp.Diff(r, tt.expected))
			}
		})
	}
}
