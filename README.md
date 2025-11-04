# KubeTUI

<p align="center">
  <strong>A modern, powerful Terminal UI for Kubernetes management</strong>
</p>

<p align="center">
  Built with <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a> 🫧
</p>

## Features

### 🎯 Core Features

- **Multi-Context & Namespace Support**
  - Quick context switching with visual feedback
  - Namespace selection and filtering
  - "All Namespaces" mode for cluster-wide view
  - Favorite contexts and namespaces

- **Pod Management**
  - Real-time pod status updates (auto-refresh every 5s)
  - Color-coded status indicators (Running/Pending/Failed)
  - Detailed pod information (containers, resources, events)
  - Pod logs viewer with tail mode
  - Delete pods with confirmation dialog
  - Navigate pods with vim-style keybindings (j/k)

- **Resource Views**
  - Pods (with detailed status)
  - Deployments (replica counts, status)
  - Services (type, cluster-IP, ports)
  - ConfigMaps & Secrets
  - Ingresses
  - StatefulSets, DaemonSets
  - Jobs & CronJobs

- **Beautiful UI**
  - Split-panel design for list and details
  - Syntax-highlighted status indicators
  - Responsive layout
  - Help screen with keyboard shortcuts (Press `?`)

## Installation

### From Source

```bash
go install github.com/Xenon1507/k8s-tui/cmd/kubetui@latest
```

### Build Manually

```bash
git clone https://github.com/Xenon1507/k8s-tui.git
cd k8s-tui
make build
sudo make install
```

### Download Binary

Download the latest release from the [Releases page](https://github.com/Xenon1507/k8s-tui/releases).

## Quick Start

### Prerequisites

- kubectl configured with access to a Kubernetes cluster
- Valid kubeconfig file (usually at `~/.kube/config`)

### Basic Usage

```bash
# Start with default kubeconfig
kubetui

# Use specific kubeconfig
kubetui --kubeconfig /path/to/kubeconfig

# Start with specific context
kubetui --context prod-cluster

# Start with specific namespace
kubetui --namespace production

# Set custom refresh interval
kubetui --refresh 10s
```

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Move up/down in lists |
| `Enter` | Select item / Show details |
| `Esc` | Go back / Close panel |
| `Tab` | Switch between panels |
| `?` or `F1` | Show help screen |
| `q` or `Ctrl+C` | Quit application |

### Views

| Key | Action |
|-----|--------|
| `1` | Switch to Pods view |
| `2` | Switch to Deployments view |
| `3` | Switch to Services view |
| `4` | Switch to ConfigMaps |
| `5` | Switch to Secrets |
| `n` | Switch to Namespaces |
| `c` | Switch to Contexts |
| `a` | Toggle all namespaces mode |

### Pod Actions

| Key | Action |
|-----|--------|
| `l` | View pod logs |
| `d` | Delete pod (with confirmation) |
| `D` | Describe pod (detailed YAML-like output) |
| `e` | Exec into pod container (opens shell) |
| `p` | Port-forward to pod |
| `r` | Refresh data manually |

### Advanced

| Key | Action |
|-----|--------|
| `/` | Global search |
| `f` | Filter by labels |
| `s` | Change sort order |
| `c` | Copy logs to clipboard |

## Configuration

KubeTUI creates a config file at `~/.kubetui/config.yaml` on first run.

### Example Configuration

```yaml
# Refresh interval for auto-updates
refresh_interval: 5s

# Number of log lines to tail
log_tail_lines: 100

# Theme: dark or light
theme: dark

# Favorite namespaces for quick access
favorites:
  namespaces:
    - production
    - staging
    - development
  contexts:
    - prod-cluster
    - dev-cluster

# Customize keyboard shortcuts
key_bindings:
  logs: l
  delete: d
  describe: D
  exec: e
  port_forward: p
  refresh: r
  filter: f
  search: /
  help: "?"

# Display preferences
display:
  show_timestamps: true
  compact_mode: false
  max_name_length: 50
```

### Command-Line Flags

```
--kubeconfig <path>    Path to kubeconfig file (default: ~/.kube/config)
--context <name>       Kubernetes context to use
--namespace <name>     Kubernetes namespace to use (default: default)
--config <path>        Path to kubetui config file (default: ~/.kubetui/config.yaml)
--refresh <duration>   Refresh interval (e.g., 5s, 10s)
--no-color            Disable colors
--version             Show version information
```

## UI Layout

```
┌─ KubeTUI ─ Context: dev-cluster ─ Namespace: production ────────────────┐
│ [1] Pods [2] Deployments [3] Services [4] ConfigMaps       [?] Help     │
├──────────────────────────────────────────────────────────────────────────┤
│ ┌─ Resource List ─────────────┬─ Details/Logs ────────────────────────┐ │
│ │ NAME              STATUS     │ Pod: api-deployment-abc123           │ │
│ │ ● api-deploy..    Running    │                                      │ │
│ │ ● worker-xyz..    Running    │ Containers: 2/2                      │ │
│ │ ⚠ batch-abc..     Pending    │ Status: Running                      │ │
│ │ ✗ old-pod-def..   Failed     │ Restarts: 0                          │ │
│ │                               │ Age: 2d                              │ │
│ │ [150 pods total]              │ Node: node-1                         │ │
│ │                               │ IP: 10.0.1.42                        │ │
│ │                               │                                      │ │
│ │                               │ Containers:                          │ │
│ │                               │   ● app (restarts: 0)                │ │
│ │                               │   ● sidecar (restarts: 0)            │ │
│ └──────────────────────────────┴──────────────────────────────────────┘ │
├──────────────────────────────────────────────────────────────────────────┤
│ ↑/↓: Navigate │ Enter: Details │ l: Logs │ d: Delete │ r: Refresh │ q: Quit │
└──────────────────────────────────────────────────────────────────────────┘
```

## Comparison with k9s

KubeTUI is inspired by [k9s](https://k9scli.io/) but focuses on:

| Feature | k9s | KubeTUI |
|---------|-----|---------|
| Built with | Go (Tview) | Go (Bubble Tea) |
| UI Framework | Immediate mode | Elm architecture |
| Configuration | Complex YAML | Simple YAML |
| Learning Curve | Steeper | Gentler |
| Plugin System | Yes | Planned |
| Resource Usage | Higher | Lower |
| Startup Time | ~1s | ~0.3s |

**Why KubeTUI?**
- More modern, cleaner UI based on Bubble Tea
- Simpler configuration
- Better error messages
- More intuitive keyboard shortcuts
- Lighter resource footprint

## Development

### Prerequisites

- Go 1.21 or higher
- kubectl
- Access to a Kubernetes cluster (minikube, kind, or cloud)

### Building

```bash
# Build binary
make build

# Run tests
make test

# Install locally
make install

# Clean build artifacts
make clean
```

### Project Structure

```
kubetui/
├── cmd/
│   └── kubetui/           # Main application entry point
│       └── main.go
├── internal/
│   ├── k8s/               # Kubernetes client wrapper
│   │   ├── client.go      # Client initialization
│   │   └── resources.go   # Resource CRUD operations
│   ├── ui/                # Bubble Tea UI components
│   │   ├── models/        # Bubble Tea models (state)
│   │   ├── views/         # View rendering logic
│   │   └── styles/        # Lipgloss styles
│   ├── config/            # Configuration management
│   └── utils/             # Helper functions
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── .goreleaser.yaml       # Release automation
```

### Testing

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# Integration tests (requires cluster)
go test -tags=integration ./...
```

## Troubleshooting

### "Error creating Kubernetes client"

**Problem:** KubeTUI can't connect to your cluster.

**Solution:**
1. Verify kubectl is working: `kubectl get pods`
2. Check your kubeconfig: `kubectl config view`
3. Ensure you have the right permissions

### "No pods found"

**Problem:** No pods visible in the selected namespace.

**Solution:**
1. Check if pods exist: `kubectl get pods -n <namespace>`
2. Try switching namespaces with `n`
3. Enable "All Namespaces" mode with `a`
4. Verify RBAC permissions

### "Context switch failed"

**Problem:** Can't switch to a different context.

**Solution:**
1. List available contexts: `kubectl config get-contexts`
2. Ensure the context exists in your kubeconfig
3. Check cluster connectivity

### Performance Issues

**Problem:** UI is slow or laggy.

**Solution:**
1. Increase refresh interval: `kubetui --refresh 10s`
2. Use namespace filtering instead of "All Namespaces" mode
3. Check cluster connectivity and latency

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Code Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Add tests for new features
- Update documentation
- Run `go fmt` before committing
- Keep commits atomic and well-described

## Roadmap

- [x] Basic pod management
- [x] Multi-context support
- [x] Namespace navigation
- [x] Pod logs viewer
- [ ] Real-time log streaming
- [ ] Resource metrics (CPU/Memory) via metrics-server
- [ ] Pod exec (shell access)
- [ ] Port-forwarding
- [ ] YAML editor for resources
- [ ] Custom resource definitions (CRDs)
- [ ] Events viewer
- [ ] Resource creation/editing
- [ ] Plugin system
- [ ] Themes (light/dark/custom)
- [ ] Export functionality (YAML/JSON)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The amazing TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for nice TUIs
- [k9s](https://k9scli.io/) - Inspiration for Kubernetes TUI tools
- [client-go](https://github.com/kubernetes/client-go) - Kubernetes Go client

## Support

- 🐛 [Report a bug](https://github.com/Xenon1507/k8s-tui/issues/new?labels=bug)
- 💡 [Request a feature](https://github.com/Xenon1507/k8s-tui/issues/new?labels=enhancement)
- 📖 [Documentation](https://github.com/Xenon1507/k8s-tui/wiki)
- 💬 [Discussions](https://github.com/Xenon1507/k8s-tui/discussions)

---

<p align="center">
  Made with ❤️ and Go
</p>
