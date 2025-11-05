package views

import (
	"fmt"
	"sort"
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
	case models.ViewDeployments:
		content = renderDeploymentsView(m)
	case models.ViewServices:
		content = renderServicesView(m)
	case models.ViewNodes:
		content = renderNodesView(m)
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
		renderTab("4", "Nodes", m.CurrentView == models.ViewNodes),
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
	// If viewing logs, show full-screen log view
	if m.CurrentPanel == models.PanelLogs {
		return renderPodLogsFullScreen(m)
	}

	if m.Loading {
		return styles.PanelStyle.Render("Loading pods...")
	}

	if m.ErrorMessage != "" {
		return styles.ErrorStyle.Render(m.ErrorMessage)
	}

	if len(m.Pods) == 0 {
		return styles.PanelStyle.Render("No pods found")
	}

	// Split view: List on left, details on right
	listView := renderPodsList(m)

	var detailView string
	switch m.CurrentPanel {
	case models.PanelDetail:
		detailView = renderPodDetail(m)
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

		// Find matching spec container for image info
		for _, specContainer := range pod.Spec.Containers {
			if specContainer.Name == container.Name {
				details = append(details, fmt.Sprintf("    Image: %s", specContainer.Image))
				if container.ImageID != "" {
					// Extract digest from ImageID (format: docker-pullable://image@sha256:...)
					imageID := container.ImageID
					if idx := strings.Index(imageID, "@"); idx != -1 {
						digest := imageID[idx+1:]
						// Truncate long digests for readability
						if len(digest) > 71 {
							digest = digest[:71] + "..."
						}
						details = append(details, fmt.Sprintf("    Digest: %s", digest))
					} else if len(imageID) > 60 {
						// Show truncated ImageID if no digest
						details = append(details, fmt.Sprintf("    ImageID: %s...", imageID[:57]))
					} else {
						details = append(details, fmt.Sprintf("    ImageID: %s", imageID))
					}
				}
				break
			}
		}
	}

	// Labels
	if len(pod.Labels) > 0 {
		details = append(details, "")
		details = append(details, styles.SubHeaderStyle.Render("Labels:"))
		for _, k := range getSortedMapKeys(pod.Labels) {
			details = append(details, fmt.Sprintf("  %s: %s", k, pod.Labels[k]))
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

// renderPodLogs renders pod logs (deprecated - use renderPodLogsFullScreen)
func renderPodLogs(m models.Model) string {
	return renderPodLogsFullScreen(m)
}

// renderPodLogsFullScreen renders pod logs in full-screen mode with scrolling
func renderPodLogsFullScreen(m models.Model) string {
	if m.SelectedPod == nil {
		return styles.PanelStyle.Render("No pod selected")
	}

	pod := m.SelectedPod

	// Get container info
	containerInfo := ""
	if len(pod.Spec.Containers) > 0 {
		containerInfo = fmt.Sprintf(" (container: %s)", pod.Spec.Containers[0].Name)
	}

	header := styles.LogHeaderStyle.Render(fmt.Sprintf("📋 Logs: %s%s", pod.Name, containerInfo))

	var content string

	if m.Loading {
		content = "\n\nLoading logs...\n\nPlease wait..."
	} else if m.ErrorMessage != "" {
		content = "\n\n" + styles.ErrorStyle.Render(m.ErrorMessage)
	} else if len(m.Logs) > 0 {
		// Calculate visible area
		availableHeight := m.Height - 8 // Account for header, footer, borders
		if availableHeight < 10 {
			availableHeight = 10
		}

		// Get the visible slice of logs
		startLine := m.LogViewOffset
		endLine := startLine + availableHeight

		if startLine >= len(m.Logs) {
			startLine = len(m.Logs) - availableHeight
			if startLine < 0 {
				startLine = 0
			}
		}

		if endLine > len(m.Logs) {
			endLine = len(m.Logs)
		}

		visibleLogs := m.Logs[startLine:endLine]
		content = "\n" + strings.Join(visibleLogs, "\n")

		// Add scroll indicators
		scrollInfo := ""
		if len(m.Logs) > availableHeight {
			scrollPercent := int(float64(startLine) / float64(len(m.Logs)-availableHeight) * 100)
			if scrollPercent < 0 {
				scrollPercent = 0
			}
			if scrollPercent > 100 {
				scrollPercent = 100
			}
			scrollInfo = fmt.Sprintf(" [%d%%]", scrollPercent)
		}

		footer := fmt.Sprintf("\n[Lines %d-%d of %d%s] ↑↓/j/k: Scroll | Esc: Close",
			startLine+1, endLine, len(m.Logs), scrollInfo)
		content += styles.HelpDescStyle.Render(footer)
	} else {
		content = "\n\nNo logs available for this pod."
	}

	// Render in full width
	fullContent := header + content

	return styles.LogStyle.
		Width(m.Width - 4).
		Height(m.Height - 6).
		Render(fullContent)
}

// renderNamespacesView renders the namespaces list with scrolling
func renderNamespacesView(m models.Model) string {
	var rows []string

	rows = append(rows, styles.SubHeaderStyle.Render("Select a namespace:"))
	rows = append(rows, "")

	// Calculate visible area
	availableHeight := m.Height - 10 // Account for header, footer, margins
	if availableHeight < 10 {
		availableHeight = 10
	}

	totalItems := len(m.Namespaces)
	if totalItems == 0 {
		return styles.PanelStyle.Render("No namespaces found")
	}

	// Calculate visible slice
	startIdx := m.ListViewOffset
	endIdx := startIdx + availableHeight
	if endIdx > totalItems {
		endIdx = totalItems
	}

	// Ensure cursor is within bounds
	if startIdx >= totalItems {
		startIdx = 0
	}

	visibleNamespaces := m.Namespaces[startIdx:endIdx]

	for i, ns := range visibleNamespaces {
		actualIdx := startIdx + i
		name := ns.Name
		age := formatDuration(time.Since(ns.CreationTimestamp.Time))

		row := fmt.Sprintf("%s (age: %s)", name, age)

		if actualIdx == m.Cursor {
			rows = append(rows, styles.TableSelectedStyle.Render("▶ "+row))
		} else {
			rows = append(rows, "  "+row)
		}
	}

	// Add scroll indicator
	if totalItems > availableHeight {
		scrollInfo := fmt.Sprintf("\n[Showing %d-%d of %d] Use ↑↓ to scroll",
			startIdx+1, endIdx, totalItems)
		rows = append(rows, styles.HelpDescStyle.Render(scrollInfo))
	}

	return styles.PanelStyle.Render(strings.Join(rows, "\n"))
}

// renderContextsView renders the contexts list with scrolling
func renderContextsView(m models.Model) string {
	var rows []string

	rows = append(rows, styles.SubHeaderStyle.Render("Select a context:"))
	rows = append(rows, "")

	// Calculate visible area
	availableHeight := m.Height - 10
	if availableHeight < 10 {
		availableHeight = 10
	}

	totalItems := len(m.Contexts)
	if totalItems == 0 {
		return styles.PanelStyle.Render("No contexts found")
	}

	// Calculate visible slice
	startIdx := m.ListViewOffset
	endIdx := startIdx + availableHeight
	if endIdx > totalItems {
		endIdx = totalItems
	}

	if startIdx >= totalItems {
		startIdx = 0
	}

	visibleContexts := m.Contexts[startIdx:endIdx]

	for i, ctx := range visibleContexts {
		actualIdx := startIdx + i
		marker := " "
		if ctx == m.CurrentContext {
			marker = "*"
		}

		row := fmt.Sprintf("%s %s", marker, ctx)

		if actualIdx == m.Cursor {
			rows = append(rows, styles.TableSelectedStyle.Render("▶ "+row))
		} else {
			rows = append(rows, "  "+row)
		}
	}

	// Add scroll indicator
	if totalItems > availableHeight {
		scrollInfo := fmt.Sprintf("\n[Showing %d-%d of %d] Use ↑↓ to scroll",
			startIdx+1, endIdx, totalItems)
		rows = append(rows, styles.HelpDescStyle.Render(scrollInfo))
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

// renderDeploymentsView renders the deployments view with detail/log panels
func renderDeploymentsView(m models.Model) string {
	// If viewing logs, show full-screen log view
	if m.CurrentPanel == models.PanelLogs {
		return renderPodLogsFullScreen(m) // Reuse the same log viewer
	}

	if m.Loading {
		return styles.PanelStyle.Render("Loading deployments...")
	}

	if m.ErrorMessage != "" {
		return styles.ErrorStyle.Render(m.ErrorMessage)
	}

	if len(m.Deployments) == 0 {
		return styles.PanelStyle.Render("No deployments found")
	}

	// Split view: List on left, details on right
	listView := renderDeploymentsList(m)

	var detailView string
	switch m.CurrentPanel {
	case models.PanelDetail:
		detailView = renderDeploymentDetail(m)
	default:
		detailView = styles.PanelStyle.Render("Select a deployment and press Enter for details, or 'l' for logs")
	}

	// Calculate available height
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

// renderDeploymentsList renders the deployments list
func renderDeploymentsList(m models.Model) string {
	var rows []string

	// Table header
	header := fmt.Sprintf("%-3s %-35s %-12s %-15s %-8s",
		"", "NAME", "READY", "UP-TO-DATE", "AGE")
	rows = append(rows, styles.TableHeaderStyle.Render(header))

	// Table rows
	for i, deploy := range m.Deployments {
		// Calculate age
		age := formatDuration(time.Since(deploy.CreationTimestamp.Time))

		// Truncate name if needed
		name := deploy.Name
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		// Calculate ready status
		ready := fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas)
		upToDate := fmt.Sprintf("%d", deploy.Status.UpdatedReplicas)

		// Status icon
		icon := "●"
		if deploy.Status.ReadyReplicas < deploy.Status.Replicas {
			icon = "⚠"
		}

		row := fmt.Sprintf("%-3s %-35s %-12s %-15s %-8s",
			icon,
			name,
			ready,
			upToDate,
			age,
		)

		if i == m.Cursor {
			rows = append(rows, styles.TableSelectedStyle.Render(row))
		} else {
			rows = append(rows, styles.TableRowStyle.Render(row))
		}
	}

	// Add count footer
	countFooter := fmt.Sprintf("\n%d deployments total", len(m.Deployments))
	rows = append(rows, styles.HelpDescStyle.Render(countFooter))

	return styles.PanelStyle.Render(strings.Join(rows, "\n"))
}

// renderDeploymentDetail renders deployment details
func renderDeploymentDetail(m models.Model) string {
	if m.SelectedDeployment == nil {
		return styles.PanelStyle.Render("No deployment selected")
	}

	deploy := m.SelectedDeployment

	var details []string
	details = append(details, styles.SubHeaderStyle.Render(fmt.Sprintf("Deployment: %s", deploy.Name)))
	details = append(details, "")
	details = append(details, fmt.Sprintf("Namespace: %s", deploy.Namespace))
	details = append(details, fmt.Sprintf("Replicas: %d", deploy.Status.Replicas))
	details = append(details, fmt.Sprintf("Ready: %d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas))
	details = append(details, fmt.Sprintf("Up-to-date: %d", deploy.Status.UpdatedReplicas))
	details = append(details, fmt.Sprintf("Available: %d", deploy.Status.AvailableReplicas))
	details = append(details, fmt.Sprintf("Age: %s", formatDuration(time.Since(deploy.CreationTimestamp.Time))))
	details = append(details, "")

	// Strategy
	details = append(details, styles.SubHeaderStyle.Render("Strategy:"))
	details = append(details, fmt.Sprintf("  Type: %s", deploy.Spec.Strategy.Type))
	if deploy.Spec.Strategy.RollingUpdate != nil {
		details = append(details, fmt.Sprintf("  Max Unavailable: %v", deploy.Spec.Strategy.RollingUpdate.MaxUnavailable))
		details = append(details, fmt.Sprintf("  Max Surge: %v", deploy.Spec.Strategy.RollingUpdate.MaxSurge))
	}
	details = append(details, "")

	// Selector
	if deploy.Spec.Selector != nil && len(deploy.Spec.Selector.MatchLabels) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Selector:"))
		for _, k := range getSortedMapKeys(deploy.Spec.Selector.MatchLabels) {
			details = append(details, fmt.Sprintf("  %s: %s", k, deploy.Spec.Selector.MatchLabels[k]))
		}
		details = append(details, "")
	}

	// Labels
	if len(deploy.Labels) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Labels:"))
		for _, k := range getSortedMapKeys(deploy.Labels) {
			details = append(details, fmt.Sprintf("  %s: %s", k, deploy.Labels[k]))
		}
		details = append(details, "")
	}

	// Containers
	if len(deploy.Spec.Template.Spec.Containers) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Containers:"))
		for _, container := range deploy.Spec.Template.Spec.Containers {
			details = append(details, fmt.Sprintf("  ● %s", container.Name))
			details = append(details, fmt.Sprintf("    Image: %s", container.Image))
			if len(container.Ports) > 0 {
				ports := ""
				for i, port := range container.Ports {
					if i > 0 {
						ports += ", "
					}
					ports += fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol)
				}
				details = append(details, fmt.Sprintf("    Ports: %s", ports))
			}
		}
	}

	// Conditions
	if len(deploy.Status.Conditions) > 0 {
		details = append(details, "")
		details = append(details, styles.SubHeaderStyle.Render("Conditions:"))
		for _, cond := range deploy.Status.Conditions {
			status := "✓"
			if cond.Status != "True" {
				status = "✗"
			}
			details = append(details, fmt.Sprintf("  %s %s: %s", status, cond.Type, cond.Status))
			if cond.Message != "" {
				msg := cond.Message
				if len(msg) > 60 {
					msg = msg[:57] + "..."
				}
				details = append(details, fmt.Sprintf("    %s", msg))
			}
		}
	}

	return styles.ActivePanelStyle.Render(strings.Join(details, "\n"))
}

// renderServicesView renders the services view with detail panel
func renderServicesView(m models.Model) string {
	if m.Loading {
		return styles.PanelStyle.Render("Loading services...")
	}

	if m.ErrorMessage != "" {
		return styles.ErrorStyle.Render(m.ErrorMessage)
	}

	if len(m.Services) == 0 {
		return styles.PanelStyle.Render("No services found")
	}

	// Split view: List on left, details on right
	listView := renderServicesList(m)

	var detailView string
	switch m.CurrentPanel {
	case models.PanelDetail:
		detailView = renderServiceDetail(m)
	default:
		detailView = styles.PanelStyle.Render("Select a service and press Enter for details")
	}

	// Calculate available height
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

// renderServicesList renders the services list
func renderServicesList(m models.Model) string {
	var rows []string

	// Table header
	header := fmt.Sprintf("%-3s %-30s %-15s %-15s %-20s %-8s",
		"", "NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "AGE")
	rows = append(rows, styles.TableHeaderStyle.Render(header))

	// Table rows
	for i, svc := range m.Services {
		// Calculate age
		age := formatDuration(time.Since(svc.CreationTimestamp.Time))

		// Truncate name if needed
		name := svc.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		// Get external IP
		externalIP := "<none>"
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			if svc.Status.LoadBalancer.Ingress[0].IP != "" {
				externalIP = svc.Status.LoadBalancer.Ingress[0].IP
			} else if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
				externalIP = svc.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		// Truncate external IP if too long
		if len(externalIP) > 18 {
			externalIP = externalIP[:15] + "..."
		}

		row := fmt.Sprintf("%-3s %-30s %-15s %-15s %-20s %-8s",
			"●",
			name,
			svc.Spec.Type,
			svc.Spec.ClusterIP,
			externalIP,
			age,
		)

		if i == m.Cursor {
			rows = append(rows, styles.TableSelectedStyle.Render(row))
		} else {
			rows = append(rows, styles.TableRowStyle.Render(row))
		}
	}

	// Add count footer
	countFooter := fmt.Sprintf("\n%d services total", len(m.Services))
	rows = append(rows, styles.HelpDescStyle.Render(countFooter))

	return styles.PanelStyle.Render(strings.Join(rows, "\n"))
}

// renderServiceDetail renders service details
func renderServiceDetail(m models.Model) string {
	if m.SelectedService == nil {
		return styles.PanelStyle.Render("No service selected")
	}

	svc := m.SelectedService

	var details []string
	details = append(details, styles.SubHeaderStyle.Render(fmt.Sprintf("Service: %s", svc.Name)))
	details = append(details, "")
	details = append(details, fmt.Sprintf("Namespace: %s", svc.Namespace))
	details = append(details, fmt.Sprintf("Type: %s", svc.Spec.Type))
	details = append(details, fmt.Sprintf("Cluster IP: %s", svc.Spec.ClusterIP))
	details = append(details, fmt.Sprintf("Age: %s", formatDuration(time.Since(svc.CreationTimestamp.Time))))
	details = append(details, "")

	// External IPs
	if len(svc.Spec.ExternalIPs) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("External IPs:"))
		for _, ip := range svc.Spec.ExternalIPs {
			details = append(details, fmt.Sprintf("  %s", ip))
		}
		details = append(details, "")
	}

	// Load Balancer Ingress
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Load Balancer Ingress:"))
		for _, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				details = append(details, fmt.Sprintf("  IP: %s", ing.IP))
			}
			if ing.Hostname != "" {
				details = append(details, fmt.Sprintf("  Hostname: %s", ing.Hostname))
			}
		}
		details = append(details, "")
	}

	// Ports
	if len(svc.Spec.Ports) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Ports:"))
		for _, port := range svc.Spec.Ports {
			portStr := fmt.Sprintf("  %s: %d", port.Name, port.Port)
			if port.TargetPort.IntVal > 0 {
				portStr += fmt.Sprintf(" -> %d", port.TargetPort.IntVal)
			} else if port.TargetPort.StrVal != "" {
				portStr += fmt.Sprintf(" -> %s", port.TargetPort.StrVal)
			}
			portStr += fmt.Sprintf("/%s", port.Protocol)
			if port.NodePort > 0 {
				portStr += fmt.Sprintf(" (NodePort: %d)", port.NodePort)
			}
			details = append(details, portStr)
		}
		details = append(details, "")
	}

	// Selector
	if len(svc.Spec.Selector) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Selector:"))
		for _, k := range getSortedMapKeys(svc.Spec.Selector) {
			details = append(details, fmt.Sprintf("  %s: %s", k, svc.Spec.Selector[k]))
		}
		details = append(details, "")
	}

	// Labels
	if len(svc.Labels) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Labels:"))
		for _, k := range getSortedMapKeys(svc.Labels) {
			details = append(details, fmt.Sprintf("  %s: %s", k, svc.Labels[k]))
		}
		details = append(details, "")
	}

	// Session Affinity
	if svc.Spec.SessionAffinity != "" {
		details = append(details, styles.SubHeaderStyle.Render("Session Affinity:"))
		details = append(details, fmt.Sprintf("  %s", svc.Spec.SessionAffinity))
		if svc.Spec.SessionAffinityConfig != nil && svc.Spec.SessionAffinityConfig.ClientIP != nil {
			if svc.Spec.SessionAffinityConfig.ClientIP.TimeoutSeconds != nil {
				details = append(details, fmt.Sprintf("  Timeout: %d seconds", *svc.Spec.SessionAffinityConfig.ClientIP.TimeoutSeconds))
			}
		}
	}

	return styles.ActivePanelStyle.Render(strings.Join(details, "\n"))
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

// getSortedMapKeys returns sorted keys from a map[string]string
// This ensures deterministic ordering when displaying labels/selectors
func getSortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// renderNodesView renders the nodes view with detail panel
func renderNodesView(m models.Model) string {
	if m.Loading {
		return styles.PanelStyle.Render("Loading nodes...")
	}

	if m.ErrorMessage != "" {
		return styles.ErrorStyle.Render(m.ErrorMessage)
	}

	if len(m.Nodes) == 0 {
		return styles.PanelStyle.Render("No nodes found")
	}

	// Split view: List on left, details on right
	listView := renderNodesList(m)

	var detailView string
	switch m.CurrentPanel {
	case models.PanelDetail:
		detailView = renderNodeDetail(m)
	default:
		detailView = styles.PanelStyle.Render("Select a node and press Enter for details")
	}

	// Calculate available height
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

// renderNodesList renders the nodes list
func renderNodesList(m models.Model) string {
	var rows []string

	// Table header
	header := fmt.Sprintf("%-3s %-30s %-12s %-8s",
		"", "NAME", "STATUS", "AGE")
	rows = append(rows, styles.TableHeaderStyle.Render(header))

	// Table rows
	for i, node := range m.Nodes {
		// Calculate age
		age := formatDuration(time.Since(node.CreationTimestamp.Time))

		// Truncate name if needed
		name := node.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		// Determine status
		status := "NotReady"
		icon := "○"
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" {
				if condition.Status == "True" {
					status = "Ready"
					icon = "●"
				}
				break
			}
		}

		row := fmt.Sprintf("%-3s %-30s %-12s %-8s",
			icon,
			name,
			status,
			age,
		)

		if i == m.Cursor {
			rows = append(rows, styles.TableSelectedStyle.Render(row))
		} else {
			rows = append(rows, styles.TableRowStyle.Render(row))
		}
	}

	// Add count footer
	countFooter := fmt.Sprintf("\n%d nodes total", len(m.Nodes))
	rows = append(rows, styles.HelpDescStyle.Render(countFooter))

	return styles.PanelStyle.Render(strings.Join(rows, "\n"))
}

// renderNodeDetail renders node details with resource usage bars
func renderNodeDetail(m models.Model) string {
	if m.SelectedNode == nil {
		return styles.PanelStyle.Render("No node selected")
	}

	node := m.SelectedNode

	var details []string
	details = append(details, styles.SubHeaderStyle.Render(fmt.Sprintf("Node: %s", node.Name)))
	details = append(details, "")

	// Status
	status := "NotReady"
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				status = "Ready"
			}
			break
		}
	}
	details = append(details, fmt.Sprintf("Status: %s", status))
	details = append(details, fmt.Sprintf("Age: %s", formatDuration(time.Since(node.CreationTimestamp.Time))))
	details = append(details, "")

	// Addresses
	if len(node.Status.Addresses) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Addresses:"))
		for _, addr := range node.Status.Addresses {
			details = append(details, fmt.Sprintf("  %s: %s", addr.Type, addr.Address))
		}
		details = append(details, "")
	}

	// Resource Capacity & Usage with Progress Bars
	cpuCap := node.Status.Allocatable.Cpu().MilliValue()
	memCap := node.Status.Allocatable.Memory().Value()

	details = append(details, styles.SubHeaderStyle.Render("Resources:"))

	// CPU
	cpuCapStr := fmt.Sprintf("%.2f cores", float64(cpuCap)/1000.0)
	details = append(details, fmt.Sprintf("  CPU Capacity: %s", cpuCapStr))

	if metrics, ok := m.NodeMetrics[node.Name]; ok && metrics != nil {
		cpuUsageStr := fmt.Sprintf("%.2f cores", float64(metrics.CPUUsage)/1000.0)
		details = append(details, fmt.Sprintf("  CPU Usage: %s (%.1f%%)", cpuUsageStr, metrics.CPUPercent))
		details = append(details, fmt.Sprintf("  %s", renderProgressBar(metrics.CPUPercent, 40)))
	} else {
		details = append(details, "  CPU Usage: metrics unavailable")
	}
	details = append(details, "")

	// Memory
	memCapStr := formatBytes(memCap)
	details = append(details, fmt.Sprintf("  Memory Capacity: %s", memCapStr))

	if metrics, ok := m.NodeMetrics[node.Name]; ok && metrics != nil {
		memUsageStr := formatBytes(metrics.MemoryUsage)
		details = append(details, fmt.Sprintf("  Memory Usage: %s (%.1f%%)", memUsageStr, metrics.MemPercent))
		details = append(details, fmt.Sprintf("  %s", renderProgressBar(metrics.MemPercent, 40)))
	} else {
		details = append(details, "  Memory Usage: metrics unavailable")
	}
	details = append(details, "")

	// Node Info
	details = append(details, styles.SubHeaderStyle.Render("System Info:"))
	details = append(details, fmt.Sprintf("  OS: %s", node.Status.NodeInfo.OSImage))
	details = append(details, fmt.Sprintf("  Kernel: %s", node.Status.NodeInfo.KernelVersion))
	details = append(details, fmt.Sprintf("  Container Runtime: %s", node.Status.NodeInfo.ContainerRuntimeVersion))
	details = append(details, fmt.Sprintf("  Kubelet: %s", node.Status.NodeInfo.KubeletVersion))
	details = append(details, "")

	// Conditions
	if len(node.Status.Conditions) > 0 {
		details = append(details, styles.SubHeaderStyle.Render("Conditions:"))
		for _, cond := range node.Status.Conditions {
			status := "✓"
			if cond.Status != "True" && cond.Status != "False" {
				status = "?"
			} else if (cond.Type == "Ready" && cond.Status != "True") ||
				(cond.Type != "Ready" && cond.Status == "True") {
				status = "✗"
			}
			details = append(details, fmt.Sprintf("  %s %s: %s", status, cond.Type, cond.Status))
		}
	}

	return styles.ActivePanelStyle.Render(strings.Join(details, "\n"))
}

// renderProgressBar creates a beautiful progress bar using LipGloss
func renderProgressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(width) * percent / 100.0)
	empty := width - filled

	// Choose color based on percentage
	var barColor lipgloss.Color
	if percent < 50 {
		barColor = lipgloss.Color("#00ff00") // Green
	} else if percent < 75 {
		barColor = lipgloss.Color("#ffff00") // Yellow
	} else if percent < 90 {
		barColor = lipgloss.Color("#ff8800") // Orange
	} else {
		barColor = lipgloss.Color("#ff0000") // Red
	}

	filledStyle := lipgloss.NewStyle().Foreground(barColor)
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("[%s] %.1f%%", bar, percent)
}

// formatBytes formats bytes in a human-readable way
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
