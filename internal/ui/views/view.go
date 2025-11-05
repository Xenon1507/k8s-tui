package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/Xenon1507/k8s-tui/internal/k8s"
	"github.com/Xenon1507/k8s-tui/internal/ui/models"
	"github.com/Xenon1507/k8s-tui/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// View renders the main view
func View(m models.Model) string {
	if m.ShowHelp {
		return renderHelp(m)
	}

	var content string

	// Render header
	header := renderHeader(m)

	// Render main content based on current view
	switch m.CurrentView {
	case models.ViewPods:
		content = renderPodsView(m)
	case models.ViewNamespaces:
		content = renderNamespacesView(m)
	case models.ViewContexts:
		content = renderContextsView(m)
	default:
		content = renderPodsView(m)
	}

	// Render footer
	footer := renderFooter(m)

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

// renderHeader renders the application header
func renderHeader(m models.Model) string {
	title := styles.HeaderStyle.Render("KubeTUI")

	contextInfo := fmt.Sprintf("Context: %s", m.CurrentContext)
	nsInfo := m.CurrentNamespace
	if m.AllNamespaces {
		nsInfo = "All Namespaces"
	}
	namespaceInfo := fmt.Sprintf("Namespace: %s", nsInfo)

	info := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.SubHeaderStyle.Render(contextInfo),
		styles.HelpSeparatorStyle.Render(" | "),
		styles.SubHeaderStyle.Render(namespaceInfo),
	)

	// View tabs
	tabs := renderTabs(m)

	headerContent := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", info),
		tabs,
	)

	return headerContent + "\n"
}

// renderTabs renders the view selection tabs
func renderTabs(m models.Model) string {
	tabs := []string{
		renderTab("1", "Pods", m.CurrentView == models.ViewPods),
		renderTab("2", "Deployments", m.CurrentView == models.ViewDeployments),
		renderTab("3", "Services", m.CurrentView == models.ViewServices),
		renderTab("n", "Namespaces", m.CurrentView == models.ViewNamespaces),
		renderTab("c", "Contexts", m.CurrentView == models.ViewContexts),
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

// renderTab renders a single tab
func renderTab(key, label string, active bool) string {
	style := lipgloss.NewStyle().Padding(0, 2)
	if active {
		style = style.Background(styles.ColorPrimary).Foreground(styles.ColorWhite).Bold(true)
	} else {
		style = style.Foreground(styles.ColorGray)
	}
	return style.Render(fmt.Sprintf("[%s] %s", key, label))
}

// renderPodsView renders the pods list view
func renderPodsView(m models.Model) string {
	if m.Loading {
		return styles.PanelStyle.Render("Loading pods...")
	}

	if m.ErrorMessage != "" {
		return styles.ErrorStyle.Render(m.ErrorMessage)
	}

	if len(m.Pods) == 0 {
		return styles.PanelStyle.Render("No pods found")
	}

	// Split view: List on left, details/logs on right
	listView := renderPodsList(m)

	var detailView string
	switch m.CurrentPanel {
	case models.PanelDetail:
		detailView = renderPodDetail(m)
	case models.PanelLogs:
		detailView = renderPodLogs(m)
	default:
		detailView = styles.PanelStyle.Render("Select a pod and press Enter for details, or 'l' for logs")
	}

	// Calculate available height (total - header - footer - margins)
	availableHeight := m.Height - 10
	if availableHeight < 20 {
		availableHeight = 20
	}

	listWidth := m.Width / 2
	detailWidth := m.Width - listWidth - 4

	listView = lipgloss.NewStyle().
		Width(listWidth).
		Height(availableHeight).
		Render(listView)

	detailView = lipgloss.NewStyle().
		Width(detailWidth).
		Height(availableHeight).
		Render(detailView)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

// renderPodsList renders the pods list
func renderPodsList(m models.Model) string {
	var rows []string

	// Table header
	header := fmt.Sprintf("%-3s %-30s %-12s %-8s %-8s %-8s",
		"", "NAME", "STATUS", "READY", "RESTARTS", "AGE")
	rows = append(rows, styles.TableHeaderStyle.Render(header))

	// Table rows
	for i, pod := range m.Pods {
		status := k8s.GetPodStatus(&pod)

		// Format age
		age := formatDuration(status.Age)

		// Truncate name if needed
		name := pod.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		// Status icon and style
		icon := styles.StatusIcon(string(pod.Status.Phase))
		statusStyle := styles.GetStatusStyle(string(pod.Status.Phase))

		row := fmt.Sprintf("%-3s %-30s %-12s %-8s %-8d %-8s",
			icon,
			name,
			statusStyle.Render(string(pod.Status.Phase)),
			status.Ready,
			status.Restarts,
			age,
		)

		if i == m.Cursor {
			rows = append(rows, styles.TableSelectedStyle.Render(row))
		} else {
			rows = append(rows, styles.TableRowStyle.Render(row))
		}
	}

	// Add count footer
	countFooter := fmt.Sprintf("\n%d pods total", len(m.Pods))
	rows = append(rows, styles.HelpDescStyle.Render(countFooter))

	return styles.PanelStyle.Render(strings.Join(rows, "\n"))
}

// renderPodDetail renders pod details
func renderPodDetail(m models.Model) string {
	if m.SelectedPod == nil {
		return styles.PanelStyle.Render("No pod selected")
	}

	pod := m.SelectedPod
	status := k8s.GetPodStatus(pod)

	var details []string
	details = append(details, styles.SubHeaderStyle.Render(fmt.Sprintf("Pod: %s", pod.Name)))
	details = append(details, "")
	details = append(details, fmt.Sprintf("Namespace: %s", pod.Namespace))
	details = append(details, fmt.Sprintf("Status: %s", status.Phase))
	details = append(details, fmt.Sprintf("Ready: %s", status.Ready))
	details = append(details, fmt.Sprintf("Restarts: %d", status.Restarts))
	details = append(details, fmt.Sprintf("Age: %s", formatDuration(status.Age)))
	details = append(details, fmt.Sprintf("Node: %s", status.Node))
	details = append(details, fmt.Sprintf("IP: %s", status.IP))
	details = append(details, "")

	// Container info
	details = append(details, styles.SubHeaderStyle.Render("Containers:"))
	for _, container := range pod.Status.ContainerStatuses {
		containerStatus := "●"
		if !container.Ready {
			containerStatus = "○"
		}
		details = append(details, fmt.Sprintf("  %s %s (restarts: %d)",
			containerStatus,
			container.Name,
			container.RestartCount,
		))
	}

	// Labels
	if len(pod.Labels) > 0 {
		details = append(details, "")
		details = append(details, styles.SubHeaderStyle.Render("Labels:"))
		for k, v := range pod.Labels {
			details = append(details, fmt.Sprintf("  %s: %s", k, v))
		}
	}

	// Conditions
	if len(status.Conditions) > 0 {
		details = append(details, "")
		details = append(details, styles.SubHeaderStyle.Render("Conditions:"))
		for _, cond := range status.Conditions {
			details = append(details, fmt.Sprintf("  %s: %s", cond.Type, cond.Status))
		}
	}

	return styles.ActivePanelStyle.Render(strings.Join(details, "\n"))
}

// renderPodLogs renders pod logs
func renderPodLogs(m models.Model) string {
	if m.SelectedPod == nil {
		return styles.PanelStyle.Render("No pod selected")
	}

	pod := m.SelectedPod

	// Get container info
	containerInfo := ""
	if len(pod.Spec.Containers) > 0 {
		containerInfo = fmt.Sprintf(" (container: %s)", pod.Spec.Containers[0].Name)
	}

	header := styles.LogHeaderStyle.Render(fmt.Sprintf("Logs: %s%s", pod.Name, containerInfo))

	var logsContent string

	if m.Loading {
		logsContent = "Loading logs...\n\nPlease wait..."
	} else if m.ErrorMessage != "" {
		logsContent = styles.ErrorStyle.Render(m.ErrorMessage)
	} else if len(m.Logs) > 0 {
		// Join logs with proper formatting
		logsContent = strings.Join(m.Logs, "\n")

		// Add helpful footer
		footer := fmt.Sprintf("\n\n[%d lines] Press 'esc' to close", len(m.Logs))
		logsContent += styles.HelpDescStyle.Render(footer)
	} else {
		logsContent = "No logs available for this pod."
	}

	return styles.ActivePanelStyle.Render(header + "\n\n" + logsContent)
}

// renderNamespacesView renders the namespaces list
func renderNamespacesView(m models.Model) string {
	var rows []string

	rows = append(rows, styles.SubHeaderStyle.Render("Select a namespace:"))
	rows = append(rows, "")

	for i, ns := range m.Namespaces {
		name := ns.Name
		age := formatDuration(time.Since(ns.CreationTimestamp.Time))

		row := fmt.Sprintf("%s (age: %s)", name, age)

		if i == m.Cursor {
			rows = append(rows, styles.TableSelectedStyle.Render("▶ "+row))
		} else {
			rows = append(rows, "  "+row)
		}
	}

	return styles.PanelStyle.Render(strings.Join(rows, "\n"))
}

// renderContextsView renders the contexts list
func renderContextsView(m models.Model) string {
	var rows []string

	rows = append(rows, styles.SubHeaderStyle.Render("Select a context:"))
	rows = append(rows, "")

	for i, ctx := range m.Contexts {
		marker := " "
		if ctx == m.CurrentContext {
			marker = "*"
		}

		row := fmt.Sprintf("%s %s", marker, ctx)

		if i == m.Cursor {
			rows = append(rows, styles.TableSelectedStyle.Render("▶ "+row))
		} else {
			rows = append(rows, "  "+row)
		}
	}

	return styles.PanelStyle.Render(strings.Join(rows, "\n"))
}

// renderFooter renders the footer with keyboard shortcuts
func renderFooter(m models.Model) string {
	if m.SuccessMessage != "" {
		return styles.SuccessStyle.Render("✓ " + m.SuccessMessage)
	}

	helps := []string{
		styles.HelpKeyStyle.Render("↑/↓") + " Navigate",
		styles.HelpKeyStyle.Render("Enter") + " Details",
		styles.HelpKeyStyle.Render("l") + " Logs",
		styles.HelpKeyStyle.Render("d") + " Delete",
		styles.HelpKeyStyle.Render("r") + " Refresh",
		styles.HelpKeyStyle.Render("?") + " Help",
		styles.HelpKeyStyle.Render("q") + " Quit",
	}

	return styles.FooterStyle.Render(strings.Join(helps, " │ "))
}

// renderHelp renders the help screen
func renderHelp(m models.Model) string {
	helps := []string{
		styles.TitleStyle.Render("KubeTUI - Keyboard Shortcuts"),
		"",
		styles.SubHeaderStyle.Render("Navigation:"),
		fmt.Sprintf("  %s, %s     Move up/down", styles.HelpKeyStyle.Render("↑/↓"), styles.HelpKeyStyle.Render("j/k")),
		fmt.Sprintf("  %s       Select item / Show details", styles.HelpKeyStyle.Render("Enter")),
		fmt.Sprintf("  %s        Go back / Close panel", styles.HelpKeyStyle.Render("Esc")),
		"",
		styles.SubHeaderStyle.Render("Views:"),
		fmt.Sprintf("  %s         Switch to Pods view", styles.HelpKeyStyle.Render("1")),
		fmt.Sprintf("  %s         Switch to Deployments view", styles.HelpKeyStyle.Render("2")),
		fmt.Sprintf("  %s         Switch to Services view", styles.HelpKeyStyle.Render("3")),
		fmt.Sprintf("  %s         Switch to Namespaces", styles.HelpKeyStyle.Render("n")),
		fmt.Sprintf("  %s         Switch to Contexts", styles.HelpKeyStyle.Render("c")),
		fmt.Sprintf("  %s         Toggle all namespaces", styles.HelpKeyStyle.Render("a")),
		"",
		styles.SubHeaderStyle.Render("Pod Actions:"),
		fmt.Sprintf("  %s         View logs", styles.HelpKeyStyle.Render("l")),
		fmt.Sprintf("  %s         Delete pod (with confirmation)", styles.HelpKeyStyle.Render("d")),
		fmt.Sprintf("  %s         Describe pod", styles.HelpKeyStyle.Render("D")),
		"",
		styles.SubHeaderStyle.Render("General:"),
		fmt.Sprintf("  %s         Refresh data", styles.HelpKeyStyle.Render("r")),
		fmt.Sprintf("  %s, %s    Show this help", styles.HelpKeyStyle.Render("?"), styles.HelpKeyStyle.Render("F1")),
		fmt.Sprintf("  %s, %s  Quit", styles.HelpKeyStyle.Render("q"), styles.HelpKeyStyle.Render("Ctrl+C")),
		"",
		styles.HelpDescStyle.Render("Press any key to close this help screen"),
	}

	helpContent := strings.Join(helps, "\n")

	// Center the help dialog
	return lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Center,
		lipgloss.Center,
		styles.DialogBoxStyle.Width(60).Render(helpContent),
	)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
