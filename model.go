package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chrisleechdev/maskoverlay-tui/internal/filepicker"
)

type step int

const (
	stepBase step = iota
	stepMask
	stepOptions
	stepRunning
	stepDone
)

// selectable output formats; index tracked by model.formatIdx.
var formats = []string{"gif", "webp", "png"}

// options-step focusable controls
const (
	focOpacity = iota
	focFormat
	focOutput
)

type model struct {
	step   step
	st     styles
	width  int
	height int

	basePicker filepicker.Model
	maskPicker filepicker.Model
	base       string
	mask       string

	opacity   float64
	formatIdx int
	output    textinput.Model
	optFocus  int

	spinner spinner.Model
	result  string
	err     error
}

func homeOrCwd() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

func initialModel() model {
	st := newStyles()
	start := homeOrCwd()

	bp := filepicker.New()
	bp.AllowedTypes = []string{".gif", ".png", ".jpg", ".jpeg", ".webp"}
	bp.CurrentDirectory = start
	bp.AutoHeight = false
	bp.Height = 14
	bp.DirAllowed = false
	bp.FileAllowed = true

	mp := filepicker.New()
	mp.AllowedTypes = []string{".png"}
	mp.CurrentDirectory = start
	mp.AutoHeight = false
	mp.Height = 14
	mp.DirAllowed = false
	mp.FileAllowed = true

	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "output path"
	ti.Width = 48

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = st.selected

	return model{
		step:       stepBase,
		st:         st,
		basePicker: bp,
		maskPicker: mp,
		opacity:    0.75,
		output:     ti,
		spinner:    sp,
	}
}

func (m model) Init() tea.Cmd {
	return m.basePicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h := msg.Height - 9
		if h < 3 {
			h = 3
		}
		m.basePicker.Height = h
		m.maskPicker.Height = h
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	case maskDoneMsg:
		m.step = stepDone
		m.result = msg.out
		m.err = msg.err
		return m, nil
	case spinner.TickMsg:
		if m.step == stepRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	switch m.step {
	case stepBase:
		return m.updateBase(msg)
	case stepMask:
		return m.updateMask(msg)
	case stepOptions:
		return m.updateOptions(msg)
	case stepDone:
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "enter", "q":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) updateBase(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "q" {
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.basePicker, cmd = m.basePicker.Update(msg)
	if ok, path := m.basePicker.DidSelectFile(msg); ok {
		m.base = path
		m.maskPicker.CurrentDirectory = filepath.Dir(path)
		m.step = stepMask
		return m, m.maskPicker.Init()
	}
	return m, cmd
}

func (m model) updateMask(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "q" {
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.maskPicker, cmd = m.maskPicker.Update(msg)
	if ok, path := m.maskPicker.DidSelectFile(msg); ok {
		m.mask = path
		m.prefillOutput()
		m.optFocus = focOpacity
		m.output.Blur()
		m.step = stepOptions
		return m, nil
	}
	return m, cmd
}

func (m model) updateOptions(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "tab", "down":
			m.optFocus = (m.optFocus + 1) % 3
			m.syncFocus()
			return m, nil
		case "shift+tab", "up":
			m.optFocus = (m.optFocus + 2) % 3
			m.syncFocus()
			return m, nil
		case "enter":
			out := expandHome(strings.TrimSpace(m.output.Value()))
			if out == "" {
				m.err = fmt.Errorf("output path is empty")
				return m, nil
			}
			m.err = nil
			m.step = stepRunning
			return m, tea.Batch(m.spinner.Tick, runMaskCmd(m.base, m.mask, out, m.opacity))
		case "left":
			switch m.optFocus {
			case focOpacity:
				m.opacity = clamp(m.opacity-0.05, 0, 1)
			case focFormat:
				m.formatIdx = (m.formatIdx + len(formats) - 1) % len(formats)
				m.applyFormatToOutput()
			}
			return m, nil
		case "right":
			switch m.optFocus {
			case focOpacity:
				m.opacity = clamp(m.opacity+0.05, 0, 1)
			case focFormat:
				m.formatIdx = (m.formatIdx + 1) % len(formats)
				m.applyFormatToOutput()
			}
			return m, nil
		case "q":
			if m.optFocus != focOutput {
				return m, tea.Quit
			}
		}
	}

	if m.optFocus == focOutput {
		var cmd tea.Cmd
		m.output, cmd = m.output.Update(msg)
		m.syncFormatFromOutput()
		return m, cmd
	}
	return m, nil
}

func (m *model) syncFocus() {
	if m.optFocus == focOutput {
		m.output.Focus()
	} else {
		m.output.Blur()
	}
}

func (m *model) prefillOutput() {
	dir := filepath.Dir(m.base)
	name := strings.TrimSuffix(filepath.Base(m.mask), filepath.Ext(m.mask))
	m.output.SetValue(filepath.Join(dir, name+"-masked."+formats[m.formatIdx]))
}

func (m *model) applyFormatToOutput() {
	v := m.output.Value()
	if v == "" {
		m.prefillOutput()
		return
	}
	m.output.SetValue(strings.TrimSuffix(v, filepath.Ext(v)) + "." + formats[m.formatIdx])
}

// syncFormatFromOutput keeps the format selector in step with a hand-typed ext.
func (m *model) syncFormatFromOutput() {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(m.output.Value()), "."))
	for i, f := range formats {
		if ext == f || (ext == "apng" && f == "png") {
			m.formatIdx = i
			return
		}
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		v = lo
	}
	if v > hi {
		v = hi
	}
	return math.Round(v*100) / 100
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, p[2:])
		}
	}
	return p
}

func (m model) View() string {
	switch m.step {
	case stepBase:
		return m.frame("Step 1/4 · Choose base (GIF or PNG)", m.basePicker.View(),
			"↑/↓ move · enter select · backspace up a dir · q quit")
	case stepMask:
		return m.frame("Step 2/4 · Choose mask (PNG)", m.maskPicker.View(),
			"↑/↓ move · enter select · backspace up a dir · q quit")
	case stepOptions:
		return m.frame("Step 3/4 · Options", m.optionsBody(),
			"tab next field · ←/→ adjust · enter render · esc quit")
	case stepRunning:
		return m.frame("Step 4/4 · Rendering", m.spinner.View()+" running magick…", "please wait")
	case stepDone:
		var body string
		if m.err != nil {
			body = m.st.errStyle.Render("✗ failed") + "\n\n" + m.err.Error()
		} else {
			body = m.st.success.Render("✓ wrote") + "  " + m.result
		}
		return m.frame("Done", body, "enter/esc quit")
	}
	return ""
}

func (m model) frame(title, body, help string) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.st.title.Render("maskoverlay  ·  "+title),
		m.st.panel.Render(body),
		m.st.help.Render(help),
	)
}

func (m model) optionsBody() string {
	const w = 20
	filled := int(m.opacity*float64(w) + 0.5)
	if filled > w {
		filled = w
	}
	bar := m.st.barFull.Render(strings.Repeat("█", filled)) +
		m.st.barEmpty.Render(strings.Repeat("░", w-filled))
	opacity := fmt.Sprintf("%s  %s  %.2f", m.focusLabel("Opacity", focOpacity), bar, m.opacity)

	parts := make([]string, len(formats))
	for i, f := range formats {
		if i == m.formatIdx {
			parts[i] = m.st.selected.Render("‹" + f + "›")
		} else {
			parts[i] = m.st.dim.Render(" " + f + " ")
		}
	}
	format := fmt.Sprintf("%s  %s", m.focusLabel("Format", focFormat), strings.Join(parts, " "))

	output := fmt.Sprintf("%s  %s", m.focusLabel("Output", focOutput), m.output.View())

	var errLine string
	if m.err != nil {
		errLine = "\n" + m.st.errStyle.Render(m.err.Error())
	}

	return lipgloss.JoinVertical(lipgloss.Left, opacity, "", format, "", output) + errLine
}

func (m model) focusLabel(text string, idx int) string {
	if m.optFocus == idx {
		return m.st.selected.Render("▸ " + text)
	}
	return m.st.label.Render("  " + text)
}
