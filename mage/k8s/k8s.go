package k8smagex

import (
	k8sx "github.com/nirantaraai/nava/pkg/k8s"
)

// Package-level convenience functions wrapping pkg/k8s.
var defaultRunner = k8sx.NewK8sRunner()

// Apply applies a configuration to a resource by filename or stdin.
func Apply(opts k8sx.ApplyOptions) error {
	return defaultRunner.Apply(opts)
}

// Delete deletes resources by filenames, stdin, resources and names, or by resources and label selector.
func Delete(opts k8sx.DeleteOptions, resourceNames ...string) error {
	return defaultRunner.Delete(opts, resourceNames...)
}

// CreateNamespace creates a namespace.
func CreateNamespace(name string) error {
	return defaultRunner.CreateNamespace(name)
}

// Get displays one or many resources.
func Get(resourceType, name, namespace string, args ...string) error {
	return defaultRunner.Get(resourceType, name, namespace, args...)
}

// Wait waits for a specific condition on one or many resources.
func Wait(opts k8sx.WaitOptions) error {
	return defaultRunner.Wait(opts)
}

// RolloutStatus displays the status of a rollout.
func RolloutStatus(resource, namespace string, timeout string) error {
	return defaultRunner.RolloutStatus(resource, namespace, timeout)
}

// CreateSecret creates a secret.
func CreateSecret(opts k8sx.CreateSecretOptions) error {
	return defaultRunner.CreateSecret(opts)
}

// Label labels a resource.
func Label(resourceType, name, namespace string, labels map[string]string, overwrite bool) error {
	return defaultRunner.Label(resourceType, name, namespace, labels, overwrite)
}

// Run executes a raw kubectl command.
func Run(args ...string) error {
	return defaultRunner.Run(args...)
}

// ConfigUseContext sets the current context in the kubeconfig file.
func ConfigUseContext(contextName string) error {
	return defaultRunner.ConfigUseContext(contextName)
}
