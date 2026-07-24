// kubectl-obi — kubectl plugin for OpenTelemetry eBPF Instrumentation (OBI).
// Install via krew: kubectl krew install obi
// Or directly: go install && mv kubectl-obi $(go env GOPATH)/bin/
package main

import (
	"fmt"
	"os"

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
	// TODO: apply the OBI DaemonSet manifest using client-go or kubectl exec
	// Manifest: https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/download/v0.10.0/obi-daemonset.yaml
	// 1. Create namespace if not exists
	// 2. Apply the DaemonSet YAML
	// 3. Wait for DaemonSet to be Ready
	fmt.Printf("TODO: apply DaemonSet from OBI %s release\n", obiVersion)
	fmt.Println("Once implemented, all pods on each node will be automatically instrumented.")
	return nil
}

func attachSidecar(deployment, namespace string) error {
	fmt.Printf("Injecting OBI %s sidecar into deployment %q (namespace %q)...\n", obiVersion, deployment, namespace)
	// TODO: patch the deployment to add the OBI sidecar container
	// 1. Get the deployment spec
	// 2. Add obi sidecar container with shareProcessNamespace: true + required caps
	// 3. Apply the patch
	// 4. Wait for rollout to complete
	fmt.Printf("TODO: patch deployment %q with OBI sidecar\n", deployment)
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
	// TODO: list OBI DaemonSet pods and their readiness
	// TODO: list workloads that have OBI sidecar injected
	// TODO: show which services are emitting spans (check OTel collector metrics)
	fmt.Println("TODO: query cluster for OBI DaemonSet and sidecar pods")
	fmt.Printf("  Namespace: %s\n", namespace)
	fmt.Printf("  All namespaces: %v\n", allNamespaces)
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
		Args:    cobra.ExactArgs(1),
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

func pullTraces(deployment, _ string, tail int, _ bool) error {
	fmt.Printf("Fetching last %d spans for %q...\n", tail, deployment)
	// TODO: query Jaeger/Tempo/OTel Collector API for spans with service.name=deployment
	// TODO: if --follow, open a streaming OTLP receiver or long-poll the backend
	fmt.Println("TODO: query OTLP backend for trace data")
	fmt.Println("Tip: set OTEL_BACKEND=http://jaeger:16686 or OTEL_BACKEND=http://tempo:3100")
	return nil
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
	// TODO: delete the DaemonSet and associated resources
	fmt.Println("TODO: delete OBI DaemonSet")
	return nil
}

func detachSidecar(deployment, namespace string) error {
	fmt.Printf("Removing OBI sidecar from deployment %q (namespace %q)...\n", deployment, namespace)
	// TODO: patch the deployment to remove the OBI sidecar container
	// TODO: trigger rollout restart
	fmt.Println("TODO: remove sidecar container and restart deployment")
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
