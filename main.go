package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultTime = 25 * time.Minute
	width       = 50
)

type model struct {
	duration  time.Duration
	remaining time.Duration
	timer     *time.Timer
	ticker    *time.Ticker
	styles    *Styles
	program   *tea.Program
}

type Styles struct {
	selected lipgloss.Style
	normal   lipgloss.Style
	light    lipgloss.Style
	medium   lipgloss.Style
	dark     lipgloss.Style
	white    lipgloss.Style
	red      lipgloss.Style
}

func NewStyles() *Styles {
	return &Styles{
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		light:    lipgloss.NewStyle().Foreground(lipgloss.Color("#999999")),
		medium:   lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")),
		dark:     lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")),
		white:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
		red:      lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type tickMsg time.Time

type startTimerMsg struct{}

func initialModel() model {
	return model{
		duration:  defaultTime,
		remaining: defaultTime,
		styles:    NewStyles(),
	}
}

func (m model) Init() tea.Cmd {
	m.resetTimer()
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left":
			if m.ticker != nil {
				m.ticker.Stop()
				m.ticker = nil
				m.formatDuration()
				m.remaining = m.duration
				m.resetTimer()
				return m, nil
			}
			if m.duration > 5*time.Minute {
				m.formatDuration()
				m.duration -= 5 * time.Minute
				m.remaining = m.duration
				m.resetTimer()
			} else {
				m.duration = 0 * time.Minute
				m.remaining = m.duration
				m.resetTimer()
			}
		case "right":
			if m.ticker != nil {
				m.ticker.Stop()
				m.ticker = nil
				m.formatDuration()
				m.remaining = m.duration
				m.resetTimer()
				return m, nil
			}
			m.formatDuration()
			m.duration += 5 * time.Minute
			m.remaining = m.duration
			m.resetTimer()
		}
	case startTimerMsg:
		if m.ticker == nil {
			m.ticker = time.NewTicker(time.Second)
			return m, func() tea.Msg {
				return tickMsg(time.Now())
			}
		}
	case tickMsg:
		if m.remaining > 0 && m.ticker != nil {
			m.remaining -= time.Second
			time.Sleep(time.Second)
			m.duration = m.remaining
			return m, func() tea.Msg {
				return tickMsg(time.Now())
			}
		} else if m.ticker != nil {
			m.ticker.Stop()
			m.ticker = nil
			m.remaining = m.duration
		}
	}

	return m, nil
}

// format mintue number to be multiple of 5
func (m *model) formatDuration() {
	m.duration = time.Duration(m.duration.Minutes()/5) * 5 * time.Minute
}

func (m *model) resetTimer() {
	if m.timer != nil {
		m.timer.Stop()
	}
	m.timer = time.AfterFunc(4*time.Second, func() {
		m.program.Send(startTimerMsg{})
	})
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(m.renderTimeline())
	s.WriteString("\n\n")
	s.WriteString(m.renderControls())
	s.WriteString("\n")

	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Render(s.String())

	return box
}

func (m model) renderTimeline() string {
	var s strings.Builder

	minutes := int(m.duration.Minutes())

	for i := 0; i < width; i++ {
		style := m.styles.dark
		distance := abs(i - width/2)
		if distance == 0 {
			style = m.styles.selected
		} else if distance <= 10 {
			style = m.styles.light
		} else if distance <= 20 {
			style = m.styles.medium
		}

		if (minutes-width/2+i)%5 == 0 {
			s.WriteString(style.Render("|"))
		} else {
			s.WriteString(style.Render("'"))
		}
	}
	s.WriteString("\n")

	var labels strings.Builder
	for i := 0; i < width; i++ {
		min := minutes - width/2 + i
		style := m.styles.dark
		distance := abs(i - width/2)
		if distance == 0 {
			style = m.styles.selected
		} else if distance <= 10 {
			style = m.styles.light
		} else if distance <= 20 {
			style = m.styles.medium
		}

		if min%5 == 0 && min >= 0 {
			label := fmt.Sprintf("%-5d", min)
			labels.WriteString(style.Render(label))
			i += len(label) - 1
		} else {
			labels.WriteString(" ")
		}
	}
	s.WriteString(labels.String())
	s.WriteString("\n")

	var arrow strings.Builder
	for i := 0; i < width/2; i++ {
		arrow.WriteString(" ")
	}
	arrow.WriteString(m.styles.selected.Render("▲"))
	s.WriteString(arrow.String())

	return s.String()
}

func (m model) renderControls() string {
	leftArrow := m.styles.red.Render("←")
	rightArrow := m.styles.red.Render("→")
	time := m.styles.selected.Render(formatTime(m.remaining))

	controls := lipgloss.JoinHorizontal(lipgloss.Center, leftArrow, "     ", time, "     ", rightArrow)

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, controls)
}

func formatTime(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}

func main() {
	m := initialModel()
	p := tea.NewProgram(&m)
	m.program = p
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
