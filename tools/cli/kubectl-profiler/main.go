// kubectl-profiler — kubectl plugin for OpenTelemetry eBPF Profiler.
// Install directly: go install && mv kubectl-profiler $(go env GOPATH)/bin/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// profilerVersion is the current opentelemetry-ebpf-profiler release (calendar-week tag).
const profilerVersion = "v0.0.202632"

func main() {
	root := &cobra.Command{
		Use:   "kubectl-profiler",
		Short: "Manage the OpenTelemetry eBPF Profiler in your cluster",
		Long: `kubectl-profiler deploys and manages the OpenTelemetry eBPF Profiler
(otelcol-ebpf-profiler) — continuous, zero-touch CPU profiling for every
process on a node without code changes or restarts.

Requires: Linux kernel 5.8+ with BTF enabled.`,
		SilenceUsage: true,
	}

	root.AddCommand(
		newAttachCmd(),
		newStatusCmd(),
		newDetachCmd(),
		newVersionCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// runKubectl runs kubectl with the given arguments, returning combined output.
// On non-zero exit, the error message includes stderr for context.
func runKubectl(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kubectl %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// runHelm runs helm with the given arguments, returning combined output.
func runHelm(args ...string) (string, error) {
	cmd := exec.Command("helm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("helm %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// attach —————————————————————————————————————————————————

func newAttachCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Deploy the eBPF Profiler as a DaemonSet via Helm",
		Long: `Deploy the OpenTelemetry eBPF Profiler into the cluster via Helm.

One profiler pod runs per node and instruments every process automatically.
No changes to application deployments are required.`,
		Example: `  # Deploy the profiler with default settings
  kubectl profiler attach

  # Deploy into a custom namespace
  kubectl profiler attach --namespace my-profiler`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return attach(namespace)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "profiler-system", "Namespace for profiler components")
	return cmd
}

func attach(namespace string) error {
	fmt.Printf("Deploying OpenTelemetry eBPF Profiler %s into namespace %q via Helm...\n", profilerVersion, namespace)

	fmt.Println("Adding OTel Helm repository...")
	if _, err := runHelm("repo", "add", "open-telemetry",
		"https://open-telemetry.github.io/opentelemetry-helm-charts"); err != nil &&
		!strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("attach: helm repo add: %w", err)
	}
	if _, err := runHelm("repo", "update", "open-telemetry"); err != nil {
		return fmt.Errorf("attach: helm repo update: %w", err)
	}

	fmt.Println("Installing profiler Helm chart...")
	out, err := runHelm("upgrade", "--install", "profiler",
		"open-telemetry/opentelemetry-collector",
		"--namespace", namespace, "--create-namespace",
	)
	if err != nil {
		return fmt.Errorf("attach: helm install: %w", err)
	}
	fmt.Print(out)

	fmt.Println("Waiting for DaemonSet rollout (timeout: 120s)...")
	dsName, err := runKubectl("get", "daemonset",
		"-l", "app.kubernetes.io/instance=profiler",
		"-n", namespace,
		"-o", "jsonpath={.items[0].metadata.name}")
	if err != nil {
		return fmt.Errorf("attach: find daemonset: %w", err)
	}
	out, err = runKubectl("rollout", "status", "daemonset/"+strings.TrimSpace(dsName), "-n", namespace, "--timeout=120s")
	if err != nil {
		return fmt.Errorf("attach: rollout status: %w", err)
	}
	fmt.Print(out)

	fmt.Printf("\nOpenTelemetry eBPF Profiler %s deployed successfully.\n", profilerVersion)
	fmt.Println("Run `kubectl profiler status` to verify the profiler is active.")
	return nil
}

// status —————————————————————————————————————————————————

func newStatusCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show profiler DaemonSet and pod status",
		Example: `  kubectl profiler status
  kubectl profiler status --all-namespaces`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStatus(namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "profiler-system", "Namespace to check")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Check all namespaces")
	return cmd
}

func showStatus(namespace string, allNamespaces bool) error {
	fmt.Println("OpenTelemetry eBPF Profiler Status")
	fmt.Println("-----------------------------------")

	nsArgs := func(base ...string) []string {
		if allNamespaces {
			return append(base, "--all-namespaces")
		}
		return append(base, "-n", namespace)
	}

	dsOut, err := runKubectl(nsArgs("get", "daemonset", "-l", "app.kubernetes.io/instance=profiler")...)
	if err != nil {
		fmt.Println("DaemonSets: none found")
	} else {
		fmt.Println("DaemonSets:")
		fmt.Print(dsOut)
	}

	fmt.Println()
	podOut, err := runKubectl(nsArgs("get", "pods", "-l", "app.kubernetes.io/instance=profiler")...)
	if err != nil {
		fmt.Println("Pods: none found")
	} else {
		fmt.Println("Pods:")
		fmt.Print(podOut)
	}

	return nil
}

// detach —————————————————————————————————————————————————

func newDetachCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Remove the eBPF Profiler Helm release from the cluster",
		Example: `  kubectl profiler detach
  kubectl profiler detach --namespace my-profiler`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return detach(namespace)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "profiler-system", "Namespace of profiler components")
	return cmd
}

func detach(namespace string) error {
	fmt.Printf("Removing profiler Helm release from namespace %q...\n", namespace)
	out, err := runHelm("uninstall", "profiler", "--namespace", namespace, "--ignore-not-found")
	if err != nil {
		return fmt.Errorf("detach: helm uninstall: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		fmt.Println("Profiler release not found (already removed).")
	} else {
		fmt.Print(out)
		fmt.Println("Profiler removed.")
	}
	return nil
}

// version ————————————————————————————————————————————————

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print kubectl-profiler and profiler version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kubectl-profiler: dev\n")
			fmt.Printf("Profiler:         %s\n", profilerVersion)
			fmt.Printf("Source:           https://github.com/open-telemetry/opentelemetry-ebpf-profiler\n")
		},
	}
}
