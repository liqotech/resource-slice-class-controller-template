// Copyright 2024-2025 The Liqo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main contains an an example main to setup and run a controller handling a ResourceSlice class.
package main

import (
	"flag"
	"os"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	cappedresources "github.com/liqotech/resource-slice-class-controller-template/examples/cappedresources"
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
	cappedQuantities := map[corev1.ResourceName]string{
		corev1.ResourceCPU:              "4",
		corev1.ResourceMemory:           "8Gi",
		corev1.ResourcePods:             "110",
		corev1.ResourceEphemeralStorage: "20Gi",
	}
	parsedQuantities, err := parseQuantities(cappedQuantities)
	if err != nil {
		klog.Errorf("unable to parse quantities: %v", err)
		os.Exit(1)
	}
	rsHandler := cappedresources.NewHandler(parsedQuantities)

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

func parseQuantities(quantities map[corev1.ResourceName]string) (corev1.ResourceList, error) {
	resources := corev1.ResourceList{}
	for name, quantity := range quantities {
		qnt, err := resource.ParseQuantity(quantity)
		if err != nil {
			return nil, err
		}
		resources[name] = qnt
	}
	return resources, nil
}
