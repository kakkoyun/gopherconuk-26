// kubectl-obi — kubectl plugin for OpenTelemetry eBPF Instrumentation (OBI).
// Install via krew: kubectl krew install obi
// Or directly: go install && mv kubectl-obi $(go env GOPATH)/bin/
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const obiVersion = "v0.10.0"

func main() {
	root := &cobra.Command{
		Use:   "kubectl-obi",
		Short: "Manage OBI (OpenTelemetry eBPF Instrumentation) in your cluster",
		Long: `kubectl-obi attaches, manages, and queries OBI (OpenTelemetry eBPF
Instrumentation) — zero-touch distributed tracing for Kubernetes workloads.

OBI uses eBPF to instrument services without code changes or restarts.
Requires: Linux kernel 5.8+ with BTF, and the six required capabilities.`,
		SilenceUsage: true,
	}

	root.AddCommand(
		newAttachCmd(),
		newStatusCmd(),
		newTracesCmd(),
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

// attach —————————————————————————————————————————————————

func newAttachCmd() *cobra.Command {
	var namespace string
	var mode string // daemonset | sidecar

	cmd := &cobra.Command{
		Use:   "attach [deployment]",
		Short: "Attach OBI instrumentation to a workload (or all workloads via DaemonSet)",
		Long: `Attach OBI to a specific deployment (sidecar mode) or to all workloads
on the node (DaemonSet mode, the default and recommended approach).

DaemonSet mode (default): deploys one OBI pod per node; instruments all pods
automatically. No changes to application deployments.

Sidecar mode: injects an OBI container into the target deployment's pods.
Requires a rollout restart.`,
		Example: `  # Attach to all workloads via DaemonSet (recommended)
  kubectl obi attach

  # Attach to a specific deployment via sidecar
  kubectl obi attach my-service --mode=sidecar --namespace=production`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if mode == "sidecar" && len(args) == 0 {
				return fmt.Errorf("sidecar mode requires a deployment name")
			}

			switch mode {
			case "daemonset":
				return attachDaemonSet(namespace)
			case "sidecar":
				return attachSidecar(args[0], namespace)
			default:
				return fmt.Errorf("unknown mode %q; choose daemonset or sidecar", mode)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "obi-system", "Namespace for OBI components")
	cmd.Flags().StringVar(&mode, "mode", "daemonset", "Deployment mode: daemonset (default) or sidecar")
	return cmd
}

func attachDaemonSet(namespace string) error {
	fmt.Printf("Deploying OBI %s DaemonSet into namespace %q...\n", obiVersion, namespace)

	// Create namespace idempotently: ignore "already exists" errors.
	_, err := runKubectl("create", "namespace", namespace)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("attach daemonset: create namespace: %w", err)
	}

	manifestURL := fmt.Sprintf(
		"https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/download/%s/obi-daemonset.yaml",
		obiVersion,
	)
	fmt.Printf("Applying manifest from %s...\n", manifestURL)
	out, err := runKubectl("apply", "-f", manifestURL, "-n", namespace)
	if err != nil {
		return fmt.Errorf("attach daemonset: apply manifest: %w", err)
	}
	fmt.Print(out)

	fmt.Println("Waiting for DaemonSet rollout (timeout: 120s)...")
	out, err = runKubectl("rollout", "status", "daemonset/obi-daemonset", "-n", namespace, "--timeout=120s")
	if err != nil {
		return fmt.Errorf("attach daemonset: rollout status: %w", err)
	}
	fmt.Print(out)

	fmt.Printf("\nOBI %s deployed successfully.\n", obiVersion)
	fmt.Println("Run `kubectl obi status` to verify instrumentation is active.")
	return nil
}

// jsonPatchOp is a single RFC 6902 JSON Patch operation.
type jsonPatchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

func attachSidecar(deployment, namespace string) error {
	fmt.Printf("Injecting OBI %s sidecar into deployment %q (namespace %q)...\n", obiVersion, deployment, namespace)

	if _, err := runKubectl("get", "deployment", deployment, "-n", namespace); err != nil {
		return fmt.Errorf("attach sidecar: get deployment: %w", err)
	}

	image := fmt.Sprintf("ghcr.io/open-telemetry/opentelemetry-ebpf-instrumentation/obi:%s", obiVersion)
	patchOps := []jsonPatchOp{
		{Op: "add", Path: "/spec/template/spec/shareProcessNamespace", Value: true},
		{Op: "add", Path: "/spec/template/spec/containers/-", Value: map[string]any{
			"name":            "obi",
			"image":           image,
			"imagePullPolicy": "IfNotPresent",
			"securityContext": map[string]any{"privileged": true},
		}},
	}
	patchJSON, err := json.Marshal(patchOps)
	if err != nil {
		return fmt.Errorf("attach sidecar: build patch: %w", err)
	}

	fmt.Println("Patching deployment with OBI sidecar...")
	out, err := runKubectl("patch", "deployment", deployment, "-n", namespace, "--type=json", "-p", string(patchJSON))
	if err != nil {
		return fmt.Errorf("attach sidecar: patch deployment: %w", err)
	}
	fmt.Print(out)

	out, err = runKubectl("rollout", "restart", "deployment/"+deployment, "-n", namespace)
	if err != nil {
		return fmt.Errorf("attach sidecar: rollout restart: %w", err)
	}
	fmt.Print(out)

	fmt.Println("Waiting for rollout to complete (timeout: 120s)...")
	out, err = runKubectl("rollout", "status", "deployment/"+deployment, "-n", namespace, "--timeout=120s")
	if err != nil {
		return fmt.Errorf("attach sidecar: rollout status: %w", err)
	}
	fmt.Print(out)

	return nil
}

// status —————————————————————————————————————————————————

func newStatusCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show OBI instrumentation status across the cluster",
		Example: `  kubectl obi status
  kubectl obi status --all-namespaces`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStatus(namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to check (default: current context namespace)")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Check all namespaces")
	return cmd
}

func showStatus(namespace string, allNamespaces bool) error {
	fmt.Println("OBI Status")
	fmt.Println("----------")

	// nsArgs appends the appropriate namespace flag(s) to a base argument slice.
	nsArgs := func(base ...string) []string {
		if allNamespaces {
			return append(base, "--all-namespaces")
		}
		if namespace != "" {
			return append(base, "-n", namespace)
		}
		return base
	}

	dsOut, err := runKubectl(nsArgs("get", "daemonset", "-l", "app=obi")...)
	if err != nil {
		fmt.Printf("DaemonSets: none found\n")
	} else {
		fmt.Println("DaemonSets:")
		fmt.Print(dsOut)
	}

	fmt.Println()
	podOut, err := runKubectl(nsArgs("get", "pods", "-l", "app=obi")...)
	if err != nil {
		fmt.Printf("Pods: none found\n")
	} else {
		fmt.Println("Pods:")
		fmt.Print(podOut)
	}

	return nil
}

// traces —————————————————————————————————————————————————

func newTracesCmd() *cobra.Command {
	var namespace string
	var tail int
	var follow bool

	cmd := &cobra.Command{
		Use:   "traces <deployment>",
		Short: "Pull recent trace data for a deployment",
		Long: `Query the OTel Collector or backend for recent traces from a deployment.
Requires OTEL_EXPORTER_OTLP_ENDPOINT to point to a reachable collector or
Jaeger instance.`,
		Args: cobra.ExactArgs(1),
		Example: `  kubectl obi traces my-service --tail=50
  kubectl obi traces my-service --follow`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pullTraces(args[0], namespace, tail, follow)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the deployment")
	cmd.Flags().IntVar(&tail, "tail", 20, "Number of recent spans to show")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow (stream) incoming spans")
	return cmd
}

func pullTraces(deployment, _ string, tail int, follow bool) error {
	backend := os.Getenv("OTEL_BACKEND")
	if backend == "" {
		backend = "http://localhost:16686"
	}

	fmt.Printf("Fetching last %d spans for %q from %s...\n", tail, deployment, backend)

	fetch := func() error {
		// #nosec G107 — URL is constructed from user-controlled env var, intentional.
		url := fmt.Sprintf("%s/api/traces?service=%s&limit=%d", backend, deployment, tail)
		resp, err := http.Get(url) //nolint:noctx
		if err != nil {
			return fmt.Errorf("pull traces: connect to %q: %w\nTip: set OTEL_BACKEND=http://<jaeger-host>:16686", backend, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("pull traces: read response: %w", err)
		}

		var result struct {
			Data []struct {
				TraceID string `json:"traceID"`
				Spans   []struct {
					OperationName string `json:"operationName"`
					Duration      int64  `json:"duration"`
					StartTime     int64  `json:"startTime"`
				} `json:"spans"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("pull traces: parse response: %w", err)
		}

		if len(result.Data) == 0 {
			fmt.Println("No traces found.")
			return nil
		}

		fmt.Printf("%-50s  %-12s  %s\n", "OPERATION", "DURATION", "START TIME")
		fmt.Println(strings.Repeat("-", 82))
		for _, trace := range result.Data {
			for _, span := range trace.Spans {
				dur := time.Duration(span.Duration) * time.Microsecond
				start := time.UnixMicro(span.StartTime).Format(time.RFC3339)
				fmt.Printf("%-50s  %-12s  %s\n", span.OperationName, dur, start)
			}
		}
		return nil
	}

	if err := fetch(); err != nil {
		return err
	}
	if !follow {
		return nil
	}

	fmt.Println("\nFollowing... (Ctrl-C to stop)")
	for {
		time.Sleep(2 * time.Second)
		if err := fetch(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}

// detach —————————————————————————————————————————————————

func newDetachCmd() *cobra.Command {
	var namespace string
	var mode string

	cmd := &cobra.Command{
		Use:   "detach [deployment]",
		Short: "Remove OBI instrumentation from a workload or the whole cluster",
		Example: `  # Remove the DaemonSet (stops instrumenting all pods)
  kubectl obi detach

  # Remove sidecar from a specific deployment
  kubectl obi detach my-service --mode=sidecar`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch mode {
			case "daemonset":
				return detachDaemonSet(namespace)
			case "sidecar":
				if len(args) == 0 {
					return fmt.Errorf("sidecar mode requires a deployment name")
				}
				return detachSidecar(args[0], namespace)
			default:
				return fmt.Errorf("unknown mode %q", mode)
			}
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "obi-system", "Namespace of OBI components")
	cmd.Flags().StringVar(&mode, "mode", "daemonset", "Mode to detach: daemonset or sidecar")
	return cmd
}

func detachDaemonSet(namespace string) error {
	fmt.Printf("Removing OBI DaemonSet from namespace %q...\n", namespace)
	out, err := runKubectl("delete", "daemonset", "obi-daemonset", "-n", namespace, "--ignore-not-found")
	if err != nil {
		return fmt.Errorf("detach daemonset: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		fmt.Println("OBI DaemonSet not found (already removed).")
	} else {
		fmt.Print(out)
		fmt.Println("OBI DaemonSet removed.")
	}
	return nil
}

func detachSidecar(deployment, namespace string) error {
	fmt.Printf("Removing OBI sidecar from deployment %q (namespace %q)...\n", deployment, namespace)

	// Strategic merge patch: $patch:delete removes the container by its merge key (name).
	// Setting shareProcessNamespace to false restores the default.
	mergePatch := `{"spec":{"template":{"spec":{"shareProcessNamespace":false,"containers":[{"name":"obi","$patch":"delete"}]}}}}`
	out, err := runKubectl("patch", "deployment", deployment, "-n", namespace, "--type=strategic", "-p", mergePatch)
	if err != nil {
		return fmt.Errorf("detach sidecar: patch deployment: %w", err)
	}
	fmt.Print(out)

	out, err = runKubectl("rollout", "restart", "deployment/"+deployment, "-n", namespace)
	if err != nil {
		return fmt.Errorf("detach sidecar: rollout restart: %w", err)
	}
	fmt.Print(out)

	fmt.Println("Waiting for rollout to complete (timeout: 120s)...")
	out, err = runKubectl("rollout", "status", "deployment/"+deployment, "-n", namespace, "--timeout=120s")
	if err != nil {
		return fmt.Errorf("detach sidecar: rollout status: %w", err)
	}
	fmt.Print(out)

	return nil
}

// version ————————————————————————————————————————————————

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print kubectl-obi and OBI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kubectl-obi: dev\n")
			fmt.Printf("OBI:         %s\n", obiVersion)
			fmt.Printf("Source:      https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation\n")
		},
	}
}
