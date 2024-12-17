package resourceslice

import (
	"context"
	"hash/fnv"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	"github.com/liqotech/resource-slice-classes/pkg/resourceslice/handler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ResourceSliceHandler implements the Handler interface for ResourceSlice
type ResourceSliceHandler struct {}

// NewResourceSliceHandler creates a new ResourceSliceHandler
func NewResourceSliceHandler() handler.Handler {
	return &ResourceSliceHandler{}
}

// Handle processes a ResourceSlice
func (h *ResourceSliceHandler) Handle(ctx context.Context, resourceSlice *authv1beta1.ResourceSlice) (ctrl.Result, error) {
	// Generate and update resources in status
	resources := h.generateResourcesFromName(resourceSlice.Name)
	resourceSlice.Status.Resources = resources

	klog.V(4).InfoS("Updated ResourceSlice status",
		"name", resourceSlice.Name,
		"namespace", resourceSlice.Namespace,
		"cpu", resources.Cpu().String(),
		"memory", resources.Memory().String(),
		"pods", resources.Pods().String())

	return ctrl.Result{}, nil
}

// generateResourcesFromName generates resource quantities based on the ResourceSlice name
func (h *ResourceSliceHandler) generateResourcesFromName(name string) corev1.ResourceList {
	// Create a hash of the name
	hash := fnv.New32a()
	hash.Write([]byte(name))
	hashVal := hash.Sum32()

	// Use the hash to generate resource quantities (between 1 and 10)
	cpuCount := (hashVal%10 + 1)
	memoryGB := (hashVal%5 + 1)

	return corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(int64(cpuCount), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(memoryGB*1024*1024*1024), resource.BinarySI),
		corev1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
	}
}
