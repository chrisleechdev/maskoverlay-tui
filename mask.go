package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// maskDoneMsg is delivered when the magick pipeline finishes (or fails).
type maskDoneMsg struct {
	out string
	err error
}

// formatFor maps the output extension to the imagemagick write target and
// whether the 1-bit GIF code path (threshold + dispose) should be used.
func formatFor(out string) (target string, gif bool) {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(out), ".")) {
	case "gif":
		return out, true
	case "webp":
		return out, false
	case "png", "apng":
		return "apng:" + out, false
	default:
		// Unrecognized or extensionless output: fall back to the most widely
		// supported animated format rather than erroring.
		return out, true
	}
}

// maskDims reads the mask's pixel dimensions (e.g. "160x160").
func maskDims(mask string) (string, error) {
	out, err := exec.Command("magick", "identify", "-format", "%wx%h", mask+"[0]").Output()
	if err != nil {
		return "", fmt.Errorf("identify mask: %w", err)
	}
	dims := strings.TrimSpace(string(out))
	if dims == "" {
		return "", fmt.Errorf("could not read mask dimensions")
	}
	return dims, nil
}

// runMask must stay a byte-for-byte mirror of the verified `maskoverlay` zsh
// pipeline (zsh/aliases/image.zsh) — the operator ordering is load-bearing and
// must not be "simplified".
func runMask(base, mask, out string, opacity float64) error {
	dims, err := maskDims(mask)
	if err != nil {
		return err
	}
	target, gif := formatFor(out)
	opac := strconv.FormatFloat(opacity, 'f', -1, 64)

	args := []string{
		base, "-coalesce",
		"-resize", dims + "^", "-gravity", "center", "-extent", dims,
		"null:", "(", mask, "-channel", "A", "-evaluate", "multiply", opac, "+channel", ")",
		"-compose", "Over", "-layers", "composite",
		"-alpha", "set",
		"null:", "(", mask, "-alpha", "extract",
	}
	if gif {
		args = append(args, "-threshold", "50%")
	}
	args = append(args, ")", "-compose", "CopyOpacity", "-layers", "composite")
	if gif {
		args = append(args, "-dispose", "Background")
	}
	args = append(args, "-loop", "0", target)

	cmd := exec.Command("magick", args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%v: %s", err, msg)
		}
		return err
	}
	return nil
}

// runMaskCmd wraps runMask as a tea.Cmd so the UI stays responsive while magick
// runs; the result arrives as a maskDoneMsg.
func runMaskCmd(base, mask, out string, opacity float64) tea.Cmd {
	return func() tea.Msg {
		return maskDoneMsg{out: out, err: runMask(base, mask, out, opacity)}
	}
}
