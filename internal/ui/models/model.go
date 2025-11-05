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
	ViewNodes
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
	Pods               []corev1.Pod
	Deployments        []appsv1.Deployment
	Services           []corev1.Service
	Nodes              []corev1.Node
	NodeMetrics        map[string]*k8s.NodeMetrics
	Namespaces         []corev1.Namespace
	Contexts           []string
	SelectedPod        *corev1.Pod
	SelectedDeployment *appsv1.Deployment
	SelectedService    *corev1.Service
	SelectedNode       *corev1.Node
	PodEvents          []corev1.Event
	Logs               []string

	// UI state (exported for view access)
	Cursor         int
	Width          int
	Height         int
	ErrorMessage   string
	SuccessMessage string
	Loading        bool
	SearchQuery    string
	SearchActive   bool // True when user is typing in search
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
	Pods             []corev1.Pod
	Err              error
	PreserveCursor   bool
	SavedCursor      int
	SavedViewOffset  int
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
	Deployments      []appsv1.Deployment
	Err              error
	PreserveCursor   bool
	SavedCursor      int
	SavedViewOffset  int
}

// ServicesLoadedMsg is sent when services are loaded
type ServicesLoadedMsg struct {
	Services         []corev1.Service
	Err              error
	PreserveCursor   bool
	SavedCursor      int
	SavedViewOffset  int
}

// NodesLoadedMsg is sent when nodes are loaded
type NodesLoadedMsg struct {
	Nodes    []corev1.Node
	Metrics  map[string]*k8s.NodeMetrics
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
		// Only auto-refresh resource views (Pods, Deployments, Services), not selection views
		shouldRefresh := m.autoRefresh &&
			time.Since(m.lastRefresh) >= m.config.RefreshInterval &&
			(m.CurrentView == ViewPods || m.CurrentView == ViewDeployments || m.CurrentView == ViewServices) &&
			m.CurrentPanel == PanelList // Don't refresh while viewing details/logs

		if shouldRefresh {
			m.lastRefresh = time.Now()

			// Save cursor position to restore after refresh
			savedCursor := m.Cursor
			savedOffset := m.ListViewOffset

			var refreshCmd tea.Cmd
			switch m.CurrentView {
			case ViewPods:
				refreshCmd = m.loadPodsPreservingPosition(savedCursor, savedOffset)
			case ViewDeployments:
				refreshCmd = m.loadDeploymentsPreservingPosition(savedCursor, savedOffset)
			case ViewServices:
				refreshCmd = m.loadServicesPreservingPosition(savedCursor, savedOffset)
			}

			return m, tea.Batch(refreshCmd, m.tickCmd())
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

		// Restore cursor position if this was a refresh
		if msg.PreserveCursor {
			m.Cursor = msg.SavedCursor
			m.ListViewOffset = msg.SavedViewOffset
			// Ensure cursor is still valid
			if m.Cursor >= len(m.Pods) {
				m.Cursor = len(m.Pods) - 1
				if m.Cursor < 0 {
					m.Cursor = 0
				}
			}
		} else {
			// Reset cursor if out of bounds (for initial loads)
			if m.Cursor >= len(m.Pods) {
				m.Cursor = 0
			}
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

		// Restore cursor position if this was a refresh
		if msg.PreserveCursor {
			m.Cursor = msg.SavedCursor
			m.ListViewOffset = msg.SavedViewOffset
			if m.Cursor >= len(m.Deployments) {
				m.Cursor = len(m.Deployments) - 1
				if m.Cursor < 0 {
					m.Cursor = 0
				}
			}
		} else {
			if m.Cursor >= len(m.Deployments) {
				m.Cursor = 0
			}
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

		// Restore cursor position if this was a refresh
		if msg.PreserveCursor {
			m.Cursor = msg.SavedCursor
			m.ListViewOffset = msg.SavedViewOffset
			if m.Cursor >= len(m.Services) {
				m.Cursor = len(m.Services) - 1
				if m.Cursor < 0 {
					m.Cursor = 0
				}
			}
		} else {
			if m.Cursor >= len(m.Services) {
				m.Cursor = 0
			}
		}
		return m, nil

	case NodesLoadedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.ErrorMessage = fmt.Sprintf("Error loading nodes: %v", msg.Err)
			return m, nil
		}
		m.Nodes = msg.Nodes
		m.NodeMetrics = msg.Metrics
		m.ErrorMessage = ""
		if m.Cursor >= len(m.Nodes) {
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

// matchesSearch checks if a string contains the search query (case-insensitive)
func matchesSearch(text, query string) bool {
	if query == "" {
		return true
	}
	return len(query) > 0 && len(text) > 0 &&
		containsIgnoreCase(text, query)
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// getFilteredCount returns the count of filtered items for the current view
func (m Model) getFilteredCount() int {
	if m.SearchQuery == "" {
		// No filtering, return full count
		switch m.CurrentView {
		case ViewPods:
			return len(m.Pods)
		case ViewDeployments:
			return len(m.Deployments)
		case ViewServices:
			return len(m.Services)
		case ViewNodes:
			return len(m.Nodes)
		case ViewNamespaces:
			return len(m.Namespaces)
		case ViewContexts:
			return len(m.Contexts)
		}
		return 0
	}

	// Count filtered items
	count := 0
	switch m.CurrentView {
	case ViewPods:
		for _, pod := range m.Pods {
			if matchesSearch(pod.Name, m.SearchQuery) ||
				matchesSearch(string(pod.Status.Phase), m.SearchQuery) ||
				matchesSearch(pod.Namespace, m.SearchQuery) {
				count++
			}
		}
	case ViewDeployments:
		for _, dep := range m.Deployments {
			if matchesSearch(dep.Name, m.SearchQuery) ||
				matchesSearch(dep.Namespace, m.SearchQuery) {
				count++
			}
		}
	case ViewServices:
		for _, svc := range m.Services {
			if matchesSearch(svc.Name, m.SearchQuery) ||
				matchesSearch(string(svc.Spec.Type), m.SearchQuery) ||
				matchesSearch(svc.Namespace, m.SearchQuery) {
				count++
			}
		}
	case ViewNodes:
		for _, node := range m.Nodes {
			if matchesSearch(node.Name, m.SearchQuery) {
				count++
			}
		}
	case ViewNamespaces:
		for _, ns := range m.Namespaces {
			if matchesSearch(ns.Name, m.SearchQuery) {
				count++
			}
		}
	case ViewContexts:
		for _, ctx := range m.Contexts {
			if matchesSearch(ctx, m.SearchQuery) {
				count++
			}
		}
	}
	return count
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
		if m.SearchActive {
			m.SearchActive = false
			m.SearchQuery = ""
			m.Cursor = 0 // Reset cursor when exiting search
			return m, nil
		}
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
		if !m.ShowHelp && !m.ShowConfirm && !m.SearchActive {
			m.Loading = true
			return m, m.loadPods()
		}
	case "/":
		if !m.ShowHelp && !m.ShowConfirm {
			// Only activate search in list views (not in detail/log panels)
			if m.CurrentPanel == PanelList || m.CurrentPanel == PanelDetail {
				m.SearchActive = true
				m.SearchQuery = ""
				// Switch to list panel when activating search
				m.CurrentPanel = PanelList
				return m, nil
			}
		}
	}

	// Handle search input
	if m.SearchActive {
		switch msg.String() {
		case "backspace":
			if len(m.SearchQuery) > 0 {
				m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
				m.Cursor = 0 // Reset cursor on query change
			}
			return m, nil
		case "enter":
			// Enter exits search mode
			m.SearchActive = false
			return m, nil
		case "up", "k":
			// Allow navigation during search on filtered results
			if m.Cursor > 0 {
				m.Cursor--
			}
			return m, nil
		case "down", "j":
			// Allow navigation during search on filtered results
			maxCursor := m.getFilteredCount() - 1
			if maxCursor < 0 {
				maxCursor = 0
			}
			if m.Cursor < maxCursor {
				m.Cursor++
			}
			return m, nil
		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				m.SearchQuery += msg.String()
				m.Cursor = 0 // Reset cursor on query change
				return m, nil
			}
			// Important: Return here to prevent falling through to other handlers
			return m, nil
		}
	}

	// Handle view switching
	if !m.ShowHelp && !m.ShowConfirm && !m.SearchActive {
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
		case "4":
			m.CurrentView = ViewNodes
			m.Cursor = 0
			m.ListViewOffset = 0
			m.Loading = true
			return m, m.loadNodes()
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
	if !m.ShowHelp && !m.ShowConfirm && !m.SearchActive {
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
			case ViewNodes:
				maxCursor = len(m.Nodes) - 1
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
			if m.CurrentView == ViewDeployments && len(m.Deployments) > 0 {
				m.SelectedDeployment = &m.Deployments[m.Cursor]
				m.CurrentPanel = PanelLogs
				m.Loading = true
				m.Logs = []string{} // Clear old logs
				m.LogViewOffset = 0 // Reset scroll position
				return m, m.loadDeploymentLogs()
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

	if m.CurrentView == ViewDeployments && len(m.Deployments) > 0 {
		m.SelectedDeployment = &m.Deployments[m.Cursor]
		m.CurrentPanel = PanelDetail
		return m, nil
	}

	if m.CurrentView == ViewServices && len(m.Services) > 0 {
		m.SelectedService = &m.Services[m.Cursor]
		m.CurrentPanel = PanelDetail
		return m, nil
	}

	if m.CurrentView == ViewNodes && len(m.Nodes) > 0 {
		m.SelectedNode = &m.Nodes[m.Cursor]
		m.CurrentPanel = PanelDetail
		return m, nil
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
		return PodsLoadedMsg{
			Pods:           pods,
			Err:            err,
			PreserveCursor: false,
		}
	}
}

// loadPodsPreservingPosition loads pods while preserving cursor position
func (m Model) loadPodsPreservingPosition(cursor int, offset int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		namespace := m.CurrentNamespace
		if m.AllNamespaces {
			namespace = ""
		}
		pods, err := m.client.GetPods(ctx, namespace)
		return PodsLoadedMsg{
			Pods:            pods,
			Err:             err,
			PreserveCursor:  true,
			SavedCursor:     cursor,
			SavedViewOffset: offset,
		}
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
		return DeploymentsLoadedMsg{
			Deployments:    deployments,
			Err:            err,
			PreserveCursor: false,
		}
	}
}

// loadDeploymentsPreservingPosition loads deployments while preserving cursor position
func (m Model) loadDeploymentsPreservingPosition(cursor int, offset int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		namespace := m.CurrentNamespace
		if m.AllNamespaces {
			namespace = ""
		}
		deployments, err := m.client.GetDeployments(ctx, namespace)
		return DeploymentsLoadedMsg{
			Deployments:     deployments,
			Err:             err,
			PreserveCursor:  true,
			SavedCursor:     cursor,
			SavedViewOffset: offset,
		}
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
		return ServicesLoadedMsg{
			Services:       services,
			Err:            err,
			PreserveCursor: false,
		}
	}
}

// loadServicesPreservingPosition loads services while preserving cursor position
func (m Model) loadServicesPreservingPosition(cursor int, offset int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		namespace := m.CurrentNamespace
		if m.AllNamespaces {
			namespace = ""
		}
		services, err := m.client.GetServices(ctx, namespace)
		return ServicesLoadedMsg{
			Services:        services,
			Err:             err,
			PreserveCursor:  true,
			SavedCursor:     cursor,
			SavedViewOffset: offset,
		}
	}
}

// loadNodes loads nodes and their metrics from the cluster
func (m Model) loadNodes() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		nodes, err := m.client.GetNodes(ctx)
		if err != nil {
			return NodesLoadedMsg{Nodes: nil, Metrics: nil, Err: err}
		}

		// Try to get metrics (may fail if metrics server is not installed)
		metrics, metricsErr := m.client.GetNodeMetrics(ctx)
		if metricsErr != nil {
			// Don't fail if metrics are unavailable, just log it
			metrics = make(map[string]*k8s.NodeMetrics)
		}

		return NodesLoadedMsg{
			Nodes:   nodes,
			Metrics: metrics,
			Err:     nil,
		}
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

// loadDeploymentLogs loads logs from a deployment's first available pod
func (m Model) loadDeploymentLogs() tea.Cmd {
	if m.SelectedDeployment == nil {
		return nil
	}

	deployment := m.SelectedDeployment
	tailLines := int64(m.config.LogTailLines)

	return func() tea.Msg {
		ctx := context.Background()

		// Build label selector from deployment selector
		labelSelector := ""
		if deployment.Spec.Selector != nil && deployment.Spec.Selector.MatchLabels != nil {
			for key, value := range deployment.Spec.Selector.MatchLabels {
				if labelSelector != "" {
					labelSelector += ","
				}
				labelSelector += fmt.Sprintf("%s=%s", key, value)
			}
		}

		if labelSelector == "" {
			return LogsLoadedMsg{Logs: nil, Err: fmt.Errorf("deployment has no label selector")}
		}

		// Get pods for this deployment
		pods, err := m.client.GetPodsByLabelSelector(ctx, deployment.Namespace, labelSelector)
		if err != nil {
			return LogsLoadedMsg{Logs: nil, Err: fmt.Errorf("failed to get deployment pods: %w", err)}
		}

		if len(pods) == 0 {
			return LogsLoadedMsg{Logs: []string{"No pods found for this deployment"}, Err: nil}
		}

		// Find first running pod
		var selectedPod *corev1.Pod
		for i := range pods {
			if pods[i].Status.Phase == corev1.PodRunning {
				selectedPod = &pods[i]
				break
			}
		}

		// If no running pod, use the first pod
		if selectedPod == nil {
			selectedPod = &pods[0]
		}

		// Get the first container name
		containerName := ""
		if len(selectedPod.Spec.Containers) > 0 {
			containerName = selectedPod.Spec.Containers[0].Name
		}

		// Get logs stream
		stream, err := m.client.GetPodLogs(ctx, selectedPod.Namespace, selectedPod.Name, containerName, tailLines, false, false)
		if err != nil {
			return LogsLoadedMsg{Logs: nil, Err: fmt.Errorf("failed to get logs from pod %s: %w", selectedPod.Name, err)}
		}
		defer stream.Close()

		// Read logs from stream
		var logs []string
		logs = append(logs, fmt.Sprintf("# Logs from deployment: %s (pod: %s)", deployment.Name, selectedPod.Name))
		logs = append(logs, "")

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

		if len(logs) <= 2 { // Only header lines
			logs = append(logs, "No logs available for this deployment.")
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
