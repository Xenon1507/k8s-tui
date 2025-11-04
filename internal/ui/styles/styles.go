package styles

import "github.com/charmbracelet/lipgloss"

// Color definitions
var (
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#04B575")
	ColorGreen     = lipgloss.Color("#04B575")
	ColorYellow    = lipgloss.Color("#FFD700")
	ColorRed       = lipgloss.Color("#FF5555")
	ColorGray      = lipgloss.Color("#626262")
	ColorWhite     = lipgloss.Color("#FFFFFF")
	ColorBlue      = lipgloss.Color("#61AFEF")
	ColorCyan      = lipgloss.Color("#56B6C2")
)

// Base styles
var (
	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPrimary).
			Padding(0, 1)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	// Status styles
	StatusRunning = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	StatusPending = lipgloss.NewStyle().
			Foreground(ColorYellow).
			Bold(true)

	StatusFailed = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	StatusDefault = lipgloss.NewStyle().
			Foreground(ColorGray)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorGray)

	TableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorPrimary).
				Padding(0, 1).
				Bold(true)

	// Panel styles
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGray).
			Padding(1, 2)

	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(1, 2)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	HelpSeparatorStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorRed)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	// Dialog styles
	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Background(lipgloss.Color("#282828"))

	DialogTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Align(lipgloss.Center)

	DialogButtonStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorPrimary).
				Padding(0, 2).
				Margin(0, 1)

	DialogButtonActiveStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Background(ColorWhite).
				Padding(0, 2).
				Margin(0, 1).
				Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorGray).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	// Log styles
	LogStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGray).
			Padding(1)

	LogHeaderStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true).
			Padding(0, 1)

	// Footer/Help bar style
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorGray).
			Padding(0, 1)
)

// GetStatusStyle returns the appropriate style for a given status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "Running", "Succeeded", "Active":
		return StatusRunning
	case "Pending", "Creating":
		return StatusPending
	case "Failed", "Error", "CrashLoopBackOff":
		return StatusFailed
	default:
		return StatusDefault
	}
}

// StatusIcon returns an icon for a given status
func StatusIcon(status string) string {
	switch status {
	case "Running", "Succeeded", "Active":
		return "●"
	case "Pending", "Creating":
		return "⚠"
	case "Failed", "Error", "CrashLoopBackOff":
		return "✗"
	default:
		return "○"
	}
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-2] + ".."
}
