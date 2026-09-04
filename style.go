package main

import "github.com/charmbracelet/lipgloss"

type styles struct {
	title    lipgloss.Style
	panel    lipgloss.Style
	label    lipgloss.Style
	help     lipgloss.Style
	errStyle lipgloss.Style
	success  lipgloss.Style
	barFull  lipgloss.Style
	barEmpty lipgloss.Style
	selected lipgloss.Style
	dim      lipgloss.Style
}

func newStyles() styles {
	accent := lipgloss.Color("212") // pink
	subtle := lipgloss.Color("240")
	green := lipgloss.Color("42")
	red := lipgloss.Color("203")

	return styles{
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(accent).Padding(0, 1),
		panel:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2),
		label:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")),
		help:     lipgloss.NewStyle().Foreground(subtle),
		errStyle: lipgloss.NewStyle().Foreground(red).Bold(true),
		success:  lipgloss.NewStyle().Foreground(green).Bold(true),
		barFull:  lipgloss.NewStyle().Foreground(accent),
		barEmpty: lipgloss.NewStyle().Foreground(subtle),
		selected: lipgloss.NewStyle().Bold(true).Foreground(accent),
		dim:      lipgloss.NewStyle().Foreground(subtle),
	}
}
