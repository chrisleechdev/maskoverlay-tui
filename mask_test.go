package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestRunMask is a guarded integration test: it exercises the real magick
// pipeline against the sample assets in ~/Downloads when they are present, and
// skips otherwise (so it stays a no-op in CI or on other machines).
func TestRunMask(t *testing.T) {
	if _, err := exec.LookPath("magick"); err != nil {
		t.Skip("magick not on PATH")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	base := filepath.Join(home, "Downloads", "giphy.gif")
	mask := filepath.Join(home, "Downloads", "one-emoji.png")
	for _, f := range []string{base, mask} {
		if _, err := os.Stat(f); err != nil {
			t.Skipf("sample asset missing: %s", f)
		}
	}

	out := filepath.Join(t.TempDir(), "sticker.gif")
	if err := runMask(base, mask, out, 0.75); err != nil {
		t.Fatalf("runMask: %v", err)
	}

	dims, err := maskDims(mask)
	if err != nil {
		t.Fatalf("maskDims: %v", err)
	}
	got, err := exec.Command("magick", "identify", "-format", "%wx%h", out+"[0]").Output()
	if err != nil {
		t.Fatalf("identify output: %v", err)
	}
	if string(got) != dims {
		t.Errorf("output dims = %q, want mask dims %q", got, dims)
	}
}
