package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeployments returns deployments in the specified namespace
func (c *Client) GetDeployments(ctx context.Context, namespace string) ([]appsv1.Deployment, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetDeployment returns a specific deployment
func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteDeployment deletes a deployment
func (c *Client) DeleteDeployment(ctx context.Context, namespace, name string) error {
	return c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetServices returns services in the specified namespace
func (c *Client) GetServices(ctx context.Context, namespace string) ([]corev1.Service, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetService returns a specific service
func (c *Client) GetService(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteService deletes a service
func (c *Client) DeleteService(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetConfigMaps returns configmaps in the specified namespace
func (c *Client) GetConfigMaps(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetConfigMap returns a specific configmap
func (c *Client) GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteConfigMap deletes a configmap
func (c *Client) DeleteConfigMap(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetSecrets returns secrets in the specified namespace
func (c *Client) GetSecrets(ctx context.Context, namespace string) ([]corev1.Secret, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetSecret returns a specific secret
func (c *Client) GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteSecret deletes a secret
func (c *Client) DeleteSecret(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetIngresses returns ingresses in the specified namespace
func (c *Client) GetIngresses(ctx context.Context, namespace string) ([]networkingv1.Ingress, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetIngress returns a specific ingress
func (c *Client) GetIngress(ctx context.Context, namespace, name string) (*networkingv1.Ingress, error) {
	return c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteIngress deletes an ingress
func (c *Client) DeleteIngress(ctx context.Context, namespace, name string) error {
	return c.clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetStatefulSets returns statefulsets in the specified namespace
func (c *Client) GetStatefulSets(ctx context.Context, namespace string) ([]appsv1.StatefulSet, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetStatefulSet returns a specific statefulset
func (c *Client) GetStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	return c.clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteStatefulSet deletes a statefulset
func (c *Client) DeleteStatefulSet(ctx context.Context, namespace, name string) error {
	return c.clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetDaemonSets returns daemonsets in the specified namespace
func (c *Client) GetDaemonSets(ctx context.Context, namespace string) ([]appsv1.DaemonSet, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetDaemonSet returns a specific daemonset
func (c *Client) GetDaemonSet(ctx context.Context, namespace, name string) (*appsv1.DaemonSet, error) {
	return c.clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteDaemonSet deletes a daemonset
func (c *Client) DeleteDaemonSet(ctx context.Context, namespace, name string) error {
	return c.clientset.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetJobs returns jobs in the specified namespace
func (c *Client) GetJobs(ctx context.Context, namespace string) ([]batchv1.Job, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetJob returns a specific job
func (c *Client) GetJob(ctx context.Context, namespace, name string) (*batchv1.Job, error) {
	return c.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteJob deletes a job
func (c *Client) DeleteJob(ctx context.Context, namespace, name string) error {
	return c.clientset.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetCronJobs returns cronjobs in the specified namespace
func (c *Client) GetCronJobs(ctx context.Context, namespace string) ([]batchv1.CronJob, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	list, err := c.clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetCronJob returns a specific cronjob
func (c *Client) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	return c.clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteCronJob deletes a cronjob
func (c *Client) DeleteCronJob(ctx context.Context, namespace, name string) error {
	return c.clientset.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
