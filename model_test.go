package main

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// isQuit reports whether cmd, when invoked, yields a tea.QuitMsg.
func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestGoBack(t *testing.T) {
	tests := []struct {
		name      string
		startStep step
		setup     func(*model)
		wantStep  step
		wantQuit  bool
		check     func(*testing.T, model)
	}{
		{
			name:      "mask step goes back to base",
			startStep: stepMask,
			wantStep:  stepBase,
			wantQuit:  false,
		},
		{
			name:      "options step goes back to mask and blurs output",
			startStep: stepOptions,
			setup: func(m *model) {
				m.output.Focus()
			},
			wantStep: stepMask,
			wantQuit: false,
			check: func(t *testing.T, m model) {
				if m.output.Focused() {
					t.Errorf("output.Focused() = true, want false")
				}
			},
		},
		{
			name:      "done step goes back to options and clears err",
			startStep: stepDone,
			setup: func(m *model) {
				m.err = errors.New("boom")
			},
			wantStep: stepOptions,
			wantQuit: false,
			check: func(t *testing.T, m model) {
				if m.err != nil {
					t.Errorf("m.err = %v, want nil", m.err)
				}
			},
		},
		{
			name:      "base step quits without changing step",
			startStep: stepBase,
			wantStep:  stepBase,
			wantQuit:  true,
		},
		{
			name:      "running step quits without changing step",
			startStep: stepRunning,
			wantStep:  stepRunning,
			wantQuit:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel()
			m.step = tt.startStep
			if tt.setup != nil {
				tt.setup(&m)
			}

			resultModel, cmd := m.goBack()

			got, ok := resultModel.(model)
			if !ok {
				t.Fatalf("goBack() returned tea.Model of type %T, want model", resultModel)
			}

			if got.step != tt.wantStep {
				t.Errorf("step = %v, want %v", got.step, tt.wantStep)
			}

			if quit := isQuit(cmd); quit != tt.wantQuit {
				t.Errorf("isQuit(cmd) = %v, want %v", quit, tt.wantQuit)
			}

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
