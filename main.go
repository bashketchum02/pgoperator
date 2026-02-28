package main

// main.go is the entry point for the operator.
//
// It does three things:
//   1. Builds a Kubernetes client from your kubeconfig
//   2. Creates a dynamic client for our CRD (since we don't have generated clientsets)
//   3. Creates and starts the controller
//
// KEY CONCEPT: client-go provides two ways to talk to the API server:
//
//   a) Typed clients (generated) — e.g., clientset.CoreV1().Pods("default").Get(...)
//      These are type-safe but require running code-generator for custom resources.
//
//   b) Dynamic client — works with any resource using unstructured.Unstructured objects.
//      Less type-safe but no code generation needed. We use this approach.
//
//   In production operators you'd typically generate typed clients, but the dynamic
//   client teaches you more about how K8s API works under the hood.

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	controller "github.com/apranto/pgoperator/pkg/controller/postgresdb"
)

func main() {
	// Initialize klog flags (--v=2 for verbose, etc.)
	klog.InitFlags(nil)

	// --kubeconfig flag: path to kubeconfig file
	// When running outside the cluster (like during development), we need this.
	// When running inside the cluster as a Pod, we use in-cluster config instead.
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "path to kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	}
	flag.Parse()

	// Build the config.
	// Try kubeconfig first (development), fall back to in-cluster config (production).
	config, err := buildConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %v", err)
	}

	// Create a standard typed clientset for built-in resources (Pods, Services, StatefulSets, etc.)
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating kubernetes client: %v", err)
	}

	// Create a dynamic client for our CRD.
	// The dynamic client can work with any resource — it returns map[string]interface{}
	// instead of typed structs. We convert to/from our types manually.
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating dynamic client: %v", err)
	}

	// Create the controller
	ctrl := controller.NewController(kubeClient, dynamicClient)

	// Set up signal handling for graceful shutdown.
	// When you Ctrl+C, this channel receives a signal, and we pass a stop channel
	// to the controller so it can drain the work queue cleanly.
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		klog.Infof("Received signal %v, shutting down", sig)
		close(stopCh)
	}()

	// Start the controller. This blocks until stopCh is closed.
	klog.Info("Starting ServeOperator controller")
	if err := ctrl.Run(stopCh); err != nil {
		klog.Fatalf("Error running controller: %v", err)
	}

	klog.Info("Controller stopped")
}

// buildConfig creates a *rest.Config from kubeconfig file or in-cluster environment.
func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		// Try the provided kubeconfig path
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err == nil {
			klog.Infof("Using kubeconfig: %s", kubeconfig)
			return config, nil
		}
		klog.Warningf("Failed to use kubeconfig %s: %v, trying in-cluster config", kubeconfig, err)
	}

	// Fall back to in-cluster config (uses service account mounted in the Pod)
	klog.Info("Using in-cluster config")
	return rest.InClusterConfig()
}
