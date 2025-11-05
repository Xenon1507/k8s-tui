package models

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Xenon1507/k8s-tui/internal/config"
	"github.com/Xenon1507/k8s-tui/internal/k8s"
	tea "github.com/charmbracelet/bubbletea"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ViewMode represents different view modes
type ViewMode int

const (
	ViewPods ViewMode = iota
	ViewDeployments
	ViewServices
	ViewConfigMaps
	ViewSecrets
	ViewIngresses
	ViewStatefulSets
	ViewDaemonSets
	ViewJobs
	ViewCronJobs
	ViewNamespaces
	ViewContexts
)

// PanelMode represents which panel is active
type PanelMode int

const (
	PanelList PanelMode = iota
	PanelDetail
	PanelLogs
	PanelHelp
)

// Model represents the main application model
type Model struct {
	// Kubernetes client
	client *k8s.Client
	config *config.Config

	// Current state (exported for access from main)
	CurrentView      ViewMode
	CurrentPanel     PanelMode
	CurrentNamespace string
	CurrentContext   string
	AllNamespaces    bool

	// Data (exported for view access)
	Pods         []corev1.Pod
	Deployments  []appsv1.Deployment
	Services     []corev1.Service
	Namespaces   []corev1.Namespace
	Contexts     []string
	SelectedPod  *corev1.Pod
	PodEvents    []corev1.Event
	Logs         []string

	// UI state (exported for view access)
	Cursor         int
	Width          int
	Height         int
	ErrorMessage   string
	SuccessMessage string
	Loading        bool
	SearchQuery    string
	FilterQuery    string
	ShowHelp       bool
	ShowConfirm    bool
	ConfirmMsg     string
	ConfirmAction  func() tea.Msg

	// Log viewer state
	LogViewOffset int // Scroll offset for log viewer

	// List viewer state
	ListViewOffset int // Scroll offset for lists (namespaces, contexts, etc.)

	// Refresh
	lastRefresh time.Time
	autoRefresh bool
}

// TickMsg is sent on each refresh interval
type TickMsg time.Time

// RefreshDataMsg signals data should be refreshed
type RefreshDataMsg struct{}

// PodsLoadedMsg is sent when pods are loaded
type PodsLoadedMsg struct {
	Pods []corev1.Pod
	Err  error
}

// NamespacesLoadedMsg is sent when namespaces are loaded
type NamespacesLoadedMsg struct{
	Namespaces []corev1.Namespace
	Err        error
}

// LogsLoadedMsg is sent when pod logs are loaded
type LogsLoadedMsg struct {
	Logs []string
	Err  error
}

// DeploymentsLoadedMsg is sent when deployments are loaded
type DeploymentsLoadedMsg struct {
	Deployments []appsv1.Deployment
	Err         error
}

// ServicesLoadedMsg is sent when services are loaded
type ServicesLoadedMsg struct {
	Services []corev1.Service
	Err      error
}

// ErrorMsg represents an error message
type ErrorMsg struct {
	Err error
}

// SuccessMsg represents a success message
type SuccessMsg struct {
	Message string
}

// NewModel creates a new model
func NewModel(client *k8s.Client, cfg *config.Config) Model {
	currentNs := "default"
	currentCtx := client.GetCurrentContext()
	contexts := client.GetContexts()

	return Model{
		client:           client,
		config:           cfg,
		CurrentView:      ViewPods,
		CurrentPanel:     PanelList,
		CurrentNamespace: currentNs,
		CurrentContext:   currentCtx,
		Contexts:         contexts,
		AllNamespaces:    false,
		Pods:             []corev1.Pod{},
		Namespaces:       []corev1.Namespace{},
		Cursor:           0,
		autoRefresh:      true,
		lastRefresh:      time.Now(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadNamespaces(),
		m.loadPods(),
		m.tickCmd(),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case TickMsg:
		if m.autoRefresh && time.Since(m.lastRefresh) >= m.config.RefreshInterval {
			m.lastRefresh = time.Now()
			return m, tea.Batch(
				m.loadPods(),
				m.tickCmd(),
			)
		}
		return m, m.tickCmd()

	case PodsLoadedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMessage = fmt.Sprintf("Error loading pods: %v", msg.Err)
			return m, nil
		}
		m.Pods = msg.Pods
		m.ErrorMessage = ""
		// Reset cursor if out of bounds
		if m.Cursor >= len(m.Pods) {
			m.Cursor = 0
		}
		return m, nil

	case NamespacesLoadedMsg:
		if msg.Err != nil {
			m.ErrorMessage = fmt.Sprintf("Error loading namespaces: %v", msg.Err)
			return m, nil
		}
		m.Namespaces = msg.Namespaces
		return m, nil

	case LogsLoadedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMessage = fmt.Sprintf("Error loading logs: %v", msg.Err)
			return m, nil
		}
		m.Logs = msg.Logs
		m.ErrorMessage = ""
		return m, nil

	case DeploymentsLoadedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMessage = fmt.Sprintf("Error loading deployments: %v", msg.Err)
			return m, nil
		}
		m.Deployments = msg.Deployments
		m.ErrorMessage = ""
		if m.Cursor >= len(m.Deployments) {
			m.Cursor = 0
		}
		return m, nil

	case ServicesLoadedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMessage = fmt.Sprintf("Error loading services: %v", msg.Err)
			return m, nil
		}
		m.Services = msg.Services
		m.ErrorMessage = ""
		if m.Cursor >= len(m.Services) {
			m.Cursor = 0
		}
		return m, nil

	case ErrorMsg:
		m.ErrorMessage = msg.Err.Error()
		return m, nil

	case SuccessMsg:
		m.SuccessMessage = msg.Message
		// Clear success message after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return SuccessMsg{Message: ""}
		})

	case RefreshDataMsg:
		m.lastRefresh = time.Now()
		return m, m.loadPods()
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear messages on any key press
	m.ErrorMessage = ""
	if m.SuccessMessage != "" {
		m.SuccessMessage = ""
	}

	// Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if !m.ShowConfirm && !m.ShowHelp {
			return m, tea.Quit
		}
	case "esc":
		if m.ShowHelp {
			m.ShowHelp = false
			return m, nil
		}
		if m.ShowConfirm {
			m.ShowConfirm = false
			return m, nil
		}
		if m.CurrentPanel != PanelList {
			m.CurrentPanel = PanelList
			return m, nil
		}
	case "?", "f1":
		m.ShowHelp = !m.ShowHelp
		return m, nil
	case "r":
		if !m.ShowHelp && !m.ShowConfirm {
			m.Loading = true
			return m, m.loadPods()
		}
	}

	// Handle view switching
	if !m.ShowHelp && !m.ShowConfirm {
		switch msg.String() {
		case "1":
			m.CurrentView = ViewPods
			m.Cursor = 0
			m.ListViewOffset = 0
			m.Loading = true
			return m, m.loadPods()
		case "2":
			m.CurrentView = ViewDeployments
			m.Cursor = 0
			m.ListViewOffset = 0
			m.Loading = true
			return m, m.loadDeployments()
		case "3":
			m.CurrentView = ViewServices
			m.Cursor = 0
			m.ListViewOffset = 0
			m.Loading = true
			return m, m.loadServices()
		case "n":
			m.CurrentView = ViewNamespaces
			m.Cursor = 0
			m.ListViewOffset = 0
			return m, nil
		case "c":
			m.CurrentView = ViewContexts
			m.Cursor = 0
			m.ListViewOffset = 0
			return m, nil
		case "a":
			m.AllNamespaces = !m.AllNamespaces
			m.Loading = true
			return m, m.loadPods()
		}
	}

	// Handle navigation
	if !m.ShowHelp && !m.ShowConfirm {
		switch msg.String() {
		case "up", "k":
			// If in log view, scroll up
			if m.CurrentPanel == PanelLogs {
				if m.LogViewOffset > 0 {
					m.LogViewOffset--
				}
				return m, nil
			}
			// Otherwise move cursor
			if m.Cursor > 0 {
				m.Cursor--
				// Adjust viewport if cursor goes above visible area
				if m.Cursor < m.ListViewOffset {
					m.ListViewOffset = m.Cursor
				}
			}
			return m, nil
		case "down", "j":
			// If in log view, scroll down
			if m.CurrentPanel == PanelLogs {
				// Calculate max scroll (total lines - visible lines)
				visibleLines := m.Height - 8 // Account for header, footer, margins
				if visibleLines < 10 {
					visibleLines = 10
				}
				maxScroll := len(m.Logs) - visibleLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.LogViewOffset < maxScroll {
					m.LogViewOffset++
				}
				return m, nil
			}
			// Otherwise move cursor in list
			maxCursor := 0
			switch m.CurrentView {
			case ViewPods:
				maxCursor = len(m.Pods) - 1
			case ViewDeployments:
				maxCursor = len(m.Deployments) - 1
			case ViewServices:
				maxCursor = len(m.Services) - 1
			case ViewNamespaces:
				maxCursor = len(m.Namespaces) - 1
			case ViewContexts:
				maxCursor = len(m.Contexts) - 1
			}
			if m.Cursor < maxCursor {
				m.Cursor++
				// Adjust viewport if cursor goes below visible area
				visibleLines := m.Height - 12 // Account for header, footer, margins
				if visibleLines < 10 {
					visibleLines = 10
				}
				if m.Cursor >= m.ListViewOffset+visibleLines {
					m.ListViewOffset = m.Cursor - visibleLines + 1
				}
			}
			return m, nil
		case "pgup":
			// Page up in log view
			if m.CurrentPanel == PanelLogs {
				pageSize := (m.Height - 8) / 2
				if pageSize < 5 {
					pageSize = 5
				}
				m.LogViewOffset -= pageSize
				if m.LogViewOffset < 0 {
					m.LogViewOffset = 0
				}
				return m, nil
			}
		case "pgdown":
			// Page down in log view
			if m.CurrentPanel == PanelLogs {
				pageSize := (m.Height - 8) / 2
				if pageSize < 5 {
					pageSize = 5
				}
				visibleLines := m.Height - 8
				maxScroll := len(m.Logs) - visibleLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.LogViewOffset += pageSize
				if m.LogViewOffset > maxScroll {
					m.LogViewOffset = maxScroll
				}
				return m, nil
			}
		case "home":
			// Jump to top in log view
			if m.CurrentPanel == PanelLogs {
				m.LogViewOffset = 0
				return m, nil
			}
		case "end":
			// Jump to bottom in log view
			if m.CurrentPanel == PanelLogs {
				visibleLines := m.Height - 8
				maxScroll := len(m.Logs) - visibleLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.LogViewOffset = maxScroll
				return m, nil
			}
		case "enter":
			return m.handleEnter()
		case "l":
			if m.CurrentView == ViewPods && len(m.Pods) > 0 {
				m.CurrentPanel = PanelLogs
				m.SelectedPod = &m.Pods[m.Cursor]
				m.Loading = true
				m.Logs = []string{} // Clear old logs
				m.LogViewOffset = 0 // Reset scroll position
				return m, m.loadLogs()
			}
		case "d":
			if m.CurrentView == ViewPods && len(m.Pods) > 0 {
				m.ShowConfirm = true
				m.ConfirmMsg = fmt.Sprintf("Delete pod %s?", m.Pods[m.Cursor].Name)
				m.ConfirmAction = m.deletePod
				return m, nil
			}
		}
	}

	return m, nil
}

// handleEnter handles the enter key
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.CurrentView == ViewPods && len(m.Pods) > 0 {
		m.SelectedPod = &m.Pods[m.Cursor]
		m.CurrentPanel = PanelDetail
		return m, m.loadPodEvents()
	}

	if m.CurrentView == ViewNamespaces && len(m.Namespaces) > 0 {
		m.CurrentNamespace = m.Namespaces[m.Cursor].Name
		m.CurrentView = ViewPods
		m.Cursor = 0
		m.Loading = true
		// Clear old data
		m.Pods = []corev1.Pod{}
		m.Deployments = []appsv1.Deployment{}
		m.Services = []corev1.Service{}
		return m, m.loadPods()
	}

	if m.CurrentView == ViewContexts && len(m.Contexts) > 0 {
		newContext := m.Contexts[m.Cursor]
		err := m.client.SwitchContext(newContext)
		if err != nil {
			m.ErrorMessage = fmt.Sprintf("Error switching context: %v", err)
			return m, nil
		}

		// Update context and reset namespace to default
		m.CurrentContext = newContext
		m.CurrentNamespace = "default"
		m.Contexts = m.client.GetContexts() // Refresh contexts list

		// Go to pods view
		m.CurrentView = ViewPods
		m.Cursor = 0
		m.Loading = true

		// Clear old data
		m.Pods = []corev1.Pod{}
		m.Namespaces = []corev1.Namespace{}
		m.Deployments = []appsv1.Deployment{}
		m.Services = []corev1.Service{}

		// Reload all data for new context
		return m, tea.Batch(
			m.loadNamespaces(),
			m.loadPods(),
		)
	}

	return m, nil
}

// loadPods loads pods from the cluster
func (m Model) loadPods() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		namespace := m.CurrentNamespace
		if m.AllNamespaces {
			namespace = ""
		}
		pods, err := m.client.GetPods(ctx, namespace)
		return PodsLoadedMsg{Pods: pods, Err: err}
	}
}

// loadNamespaces loads namespaces from the cluster
func (m Model) loadNamespaces() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		namespaces, err := m.client.GetNamespaces(ctx)
		return NamespacesLoadedMsg{Namespaces: namespaces, Err: err}
	}
}

// loadDeployments loads deployments from the cluster
func (m Model) loadDeployments() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		namespace := m.CurrentNamespace
		if m.AllNamespaces {
			namespace = ""
		}
		deployments, err := m.client.GetDeployments(ctx, namespace)
		return DeploymentsLoadedMsg{Deployments: deployments, Err: err}
	}
}

// loadServices loads services from the cluster
func (m Model) loadServices() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		namespace := m.CurrentNamespace
		if m.AllNamespaces {
			namespace = ""
		}
		services, err := m.client.GetServices(ctx, namespace)
		return ServicesLoadedMsg{Services: services, Err: err}
	}
}

// loadPodEvents loads events for the selected pod
func (m Model) loadPodEvents() tea.Cmd {
	if m.SelectedPod == nil {
		return nil
	}
	return func() tea.Msg {
		ctx := context.Background()
		events, err := m.client.GetPodEvents(ctx, m.SelectedPod.Namespace, m.SelectedPod.Name)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		m.PodEvents = events
		return nil
	}
}

// loadLogs loads logs for the selected pod
func (m Model) loadLogs() tea.Cmd {
	if m.SelectedPod == nil {
		return nil
	}

	pod := m.SelectedPod
	tailLines := int64(m.config.LogTailLines)

	return func() tea.Msg {
		ctx := context.Background()

		// Get the first container name
		containerName := ""
		if len(pod.Spec.Containers) > 0 {
			containerName = pod.Spec.Containers[0].Name
		}

		// Get logs stream
		stream, err := m.client.GetPodLogs(ctx, pod.Namespace, pod.Name, containerName, tailLines, false, false)
		if err != nil {
			return LogsLoadedMsg{Logs: nil, Err: fmt.Errorf("failed to get logs: %w", err)}
		}
		defer stream.Close()

		// Read logs from stream
		var logs []string
		scanner := bufio.NewScanner(stream)

		// Limit to prevent memory issues with huge logs
		maxLines := 1000
		lineCount := 0

		for scanner.Scan() && lineCount < maxLines {
			line := scanner.Text()
			logs = append(logs, line)
			lineCount++
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			return LogsLoadedMsg{Logs: logs, Err: fmt.Errorf("error reading logs: %w", err)}
		}

		if lineCount >= maxLines {
			logs = append(logs, fmt.Sprintf("\n... (showing first %d lines, use kubectl for full logs)", maxLines))
		}

		if len(logs) == 0 {
			logs = []string{"No logs available for this pod."}
		}

		return LogsLoadedMsg{Logs: logs, Err: nil}
	}
}

// deletePod deletes the currently selected pod
func (m Model) deletePod() tea.Msg {
	if m.SelectedPod == nil && len(m.Pods) == 0 {
		return ErrorMsg{Err: fmt.Errorf("no pod selected")}
	}

	pod := &m.Pods[m.Cursor]
	ctx := context.Background()
	err := m.client.DeletePod(ctx, pod.Namespace, pod.Name)
	if err != nil {
		return ErrorMsg{Err: fmt.Errorf("failed to delete pod: %w", err)}
	}

	return SuccessMsg{Message: fmt.Sprintf("Pod %s deleted", pod.Name)}
}

// tickCmd creates a tick command for auto-refresh
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// View renders the view (required by Bubble Tea)
func (m Model) View() string {
	// Import would cause circular dependency, so we'll handle this differently
	// The view rendering is done in the views package
	return ""
}
