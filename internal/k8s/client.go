package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Client wraps the Kubernetes client with additional functionality
type Client struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsv1beta1.Clientset
	config        *rest.Config
	kubeconfig    string
	rawConfig     *api.Config
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string, context string) (*Client, error) {
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Load kubeconfig
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	configOverrides := &clientcmd.ConfigOverrides{}

	if context != "" {
		configOverrides.CurrentContext = context
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load raw kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Try to create metrics client (optional)
	metricsClient, _ := metricsv1beta1.NewForConfig(config)

	return &Client{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
		kubeconfig:    kubeconfigPath,
		rawConfig:     &rawConfig,
	}, nil
}

// GetContexts returns all available contexts
func (c *Client) GetContexts() []string {
	contexts := make([]string, 0, len(c.rawConfig.Contexts))
	for name := range c.rawConfig.Contexts {
		contexts = append(contexts, name)
	}
	return contexts
}

// GetCurrentContext returns the current context name
func (c *Client) GetCurrentContext() string {
	return c.rawConfig.CurrentContext
}

// SwitchContext switches to a different context
func (c *Client) SwitchContext(context string) error {
	newClient, err := NewClient(c.kubeconfig, context)
	if err != nil {
		return err
	}
	*c = *newClient
	return nil
}

// GetNamespaces returns all namespaces
func (c *Client) GetNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	list, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPods returns pods in the specified namespace
func (c *Client) GetPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPodsByLabelSelector returns pods matching the label selector
func (c *Client) GetPodsByLabelSelector(ctx context.Context, namespace string, labelSelector string) ([]corev1.Pod, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPod returns a specific pod
func (c *Client) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeletePod deletes a pod
func (c *Client) DeletePod(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetPodLogs returns logs for a pod
func (c *Client) GetPodLogs(ctx context.Context, namespace, name, container string, tailLines int64, follow bool, previous bool) (io.ReadCloser, error) {
	opts := &corev1.PodLogOptions{
		Container: container,
		Follow:    follow,
		Previous:  previous,
	}

	if tailLines > 0 {
		opts.TailLines = &tailLines
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
	return req.Stream(ctx)
}

// GetEvents returns events for a namespace
func (c *Client) GetEvents(ctx context.Context, namespace string) ([]corev1.Event, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPodEvents returns events for a specific pod
func (c *Client) GetPodEvents(ctx context.Context, namespace, podName string) ([]corev1.Event, error) {
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName)
	list, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// PodStatus represents the status of a pod
type PodStatus struct {
	Phase      string
	Ready      string
	Restarts   int32
	Age        time.Duration
	Node       string
	IP         string
	Conditions []corev1.PodCondition
}

// GetPodStatus returns a structured pod status
func GetPodStatus(pod *corev1.Pod) PodStatus {
	status := PodStatus{
		Phase:      string(pod.Status.Phase),
		Node:       pod.Spec.NodeName,
		IP:         pod.Status.PodIP,
		Conditions: pod.Status.Conditions,
	}

	// Calculate ready containers
	readyCount := 0
	totalCount := len(pod.Status.ContainerStatuses)
	var restarts int32 = 0

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyCount++
		}
		restarts += cs.RestartCount
	}

	status.Ready = fmt.Sprintf("%d/%d", readyCount, totalCount)
	status.Restarts = restarts

	// Calculate age
	status.Age = time.Since(pod.CreationTimestamp.Time)

	return status
}

// GetStatusColor returns a color name based on pod status
func GetStatusColor(phase string, ready string) string {
	if phase == "Running" && ready != "0" {
		readyParts := ready
		if readyParts[0] == readyParts[len(readyParts)-1] && readyParts[0] != '0' {
			return "green"
		}
		return "yellow"
	}

	switch phase {
	case "Succeeded":
		return "green"
	case "Failed", "CrashLoopBackOff", "Error":
		return "red"
	case "Pending":
		return "yellow"
	default:
		return "gray"
	}
}
