package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Xenon1507/k8s-tui/internal/config"
	"github.com/Xenon1507/k8s-tui/internal/k8s"
	"github.com/Xenon1507/k8s-tui/internal/ui/models"
	"github.com/Xenon1507/k8s-tui/internal/ui/views"
	tea "github.com/charmbracelet/bubbletea"
)

// ModelWrapper wraps the model to provide the View method
type ModelWrapper struct {
	models.Model
}

// Init initializes the wrapper
func (w *ModelWrapper) Init() tea.Cmd {
	return w.Model.Init()
}

// Update updates the wrapper
func (w *ModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := w.Model.Update(msg)
	w.Model = newModel.(models.Model)
	return w, cmd
}

// View renders the view
func (w *ModelWrapper) View() string {
	return views.View(w.Model)
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse command-line flags
	var (
		kubeconfigPath = flag.String("kubeconfig", "", "Path to kubeconfig file (default: ~/.kube/config)")
		context        = flag.String("context", "", "Kubernetes context to use")
		namespace      = flag.String("namespace", "default", "Kubernetes namespace to use")
		configPath     = flag.String("config", "", "Path to kubetui config file (default: ~/.kubetui/config.yaml)")
		refreshStr     = flag.String("refresh", "", "Refresh interval (e.g., 5s, 10s)")
		noColor        = flag.Bool("no-color", false, "Disable colors")
		showVersion    = flag.Bool("version", false, "Show version information")
	)

	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("kubetui version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
		os.Exit(0)
	}

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override refresh interval if provided
	if *refreshStr != "" {
		duration, err := time.ParseDuration(*refreshStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid refresh interval: %v\n", err)
			os.Exit(1)
		}
		cfg.RefreshInterval = duration
	}

	// Create Kubernetes client
	client, err := k8s.NewClient(*kubeconfigPath, *context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nPlease ensure:\n")
		fmt.Fprintf(os.Stderr, "  1. kubectl is configured correctly\n")
		fmt.Fprintf(os.Stderr, "  2. You have access to a Kubernetes cluster\n")
		fmt.Fprintf(os.Stderr, "  3. Your kubeconfig file is valid\n")
		os.Exit(1)
	}

	// Create model
	model := models.NewModel(client, cfg)

	// Override namespace if provided
	if *namespace != "" && *namespace != "default" {
		model.CurrentNamespace = *namespace
	}

	// Disable colors if requested
	if *noColor {
		// This would require updating the styles package
		// For now, we'll just note the flag
	}

	// Create a wrapper that implements the View method
	wrapper := &ModelWrapper{Model: model}

	// Create program
	p := tea.NewProgram(
		wrapper,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
