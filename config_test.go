package main

import (
	"testing"
)

func TestReadConfig(t *testing.T) {
	for _, test := range []struct {
		fn          string
		expectError bool
	}{
		{"testdata/sample-config.toml", false},
		{"testdata/no-such-config.toml", true},
		{"testdata/missing-templates-config.toml", true},
		{"testdata/missing-passwd-config.toml", true},
	} {
		t.Run(test.fn, func(t *testing.T) {
			_, err := ReadConfig(test.fn)
			if err == nil && test.expectError {
				t.Errorf("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error %#v", err)
			}
		})
	}
}
