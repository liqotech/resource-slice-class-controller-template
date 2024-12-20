// Package main contains an an example main to setup and run a controller handling a ResourceSlice class
package main

import (
	"flag"
	"os"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	examplehandler "github.com/liqotech/resource-slice-class-controller-template/example/resourceslice"
	"github.com/liqotech/resource-slice-class-controller-template/pkg/controller"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(authv1beta1.AddToScheme(scheme))
	// Add custom resource scheme here when you have CRDs
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool
	var className string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&className, "class-name", "", "The name of the class to handle")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	klog.InitFlags(nil)

	flag.Parse()

	if className == "" {
		klog.Error("class-name is required")
		os.Exit(1)
	}

	ctrl.SetLogger(klog.NewKlogr())
	ctx := ctrl.SetupSignalHandler()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "resource-slice-classes-leader-election",
	})
	if err != nil {
		klog.Errorf("unable to start manager: %v", err)
		os.Exit(1)
	}

	// Create the handler
	rsHandler := examplehandler.NewHandler()

	if err = controller.NewResourceSliceReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("resource-slice-controller"),
		className,
		rsHandler,
	).SetupWithManager(mgr); err != nil {
		klog.Errorf("unable to setup controller: %v", err)
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Errorf("unable to set up health check: %v", err)
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Errorf("unable to set up ready check: %v", err)
		os.Exit(1)
	}

	klog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		klog.Errorf("unable to start controller: %v", err)
		os.Exit(1)
	}
}
