// Package resourceslice contains an example handler for ResourceSlice.
package resourceslice

import (
	"context"
	"hash/fnv"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	rshandler "github.com/liqotech/resource-slice-class-controller-template/pkg/resourceslice/handler"
)

// Handler implements the Handler interface for ResourceSlice.
type Handler struct{}

// NewHandler creates a new ResourceSliceHandler.
func NewHandler() rshandler.Handler {
	return &Handler{}
}

// Handle processes a ResourceSlice.
func (h *Handler) Handle(_ context.Context, resourceSlice *authv1beta1.ResourceSlice) (ctrl.Result, error) {
	// Generate and update resources in status
	resources, err := h.generateResourcesFromName(resourceSlice.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	resourceSlice.Status.Resources = resources

	klog.V(4).InfoS("Updated ResourceSlice status",
		"name", resourceSlice.Name,
		"namespace", resourceSlice.Namespace,
		"cpu", resources.Cpu().String(),
		"memory", resources.Memory().String(),
		"pods", resources.Pods().String())

	return ctrl.Result{}, nil
}

// generateResourcesFromName generates resource quantities based on the ResourceSlice name.
func (h *Handler) generateResourcesFromName(name string) (corev1.ResourceList, error) {
	// Create a hash of the name
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(name)); err != nil {
		return nil, err
	}
	hashVal := hash.Sum32()

	// Use the hash to generate resource quantities (between 1 and 10)
	cpuCount := (hashVal%10 + 1)
	memoryGB := (hashVal%5 + 1)

	return corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(int64(cpuCount), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(memoryGB*1024*1024*1024), resource.BinarySI),
		corev1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
	}, nil
}
