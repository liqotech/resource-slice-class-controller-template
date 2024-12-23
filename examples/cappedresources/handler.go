package cappedresources

import (
	"context"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	"github.com/liqotech/liqo/pkg/liqo-controller-manager/authentication"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rshandler "github.com/liqotech/resource-slice-class-controller-template/pkg/resourceslice/handler"
)

// Handler implements the Handler interface for ResourceSlice.
type Handler struct {
	capResources corev1.ResourceList
}

// NewHandler creates a new capped resources handler.
func NewHandler(maxResources corev1.ResourceList) rshandler.Handler {
	return &Handler{
		capResources: maxResources,
	}
}

// Handle processes the ResourceSlice.
func (h *Handler) Handle(_ context.Context, resourceSlice *authv1beta1.ResourceSlice) (ctrl.Result, error) {
	// Generate and update resources in status
	resources := h.getCappedResources(resourceSlice.Spec.Resources)
	resourceSlice.Status.Resources = resources

	klog.InfoS("Processed ResourceSlice resources",
		"name", resourceSlice.Name,
		"namespace", resourceSlice.Namespace,
		"cpu", resources.Cpu().String(),
		"memory", resources.Memory().String(),
		"pods", resources.Pods().String(),
		"ephemeral-storage", resources.StorageEphemeral().String())

	// Ensure the "Resources" condition is set.
	authentication.EnsureCondition(
		resourceSlice,
		authv1beta1.ResourceSliceConditionTypeResources,
		authv1beta1.ResourceSliceConditionAccepted,
		"ResourceSliceResourcesAccepted",
		"ResourceSlice resources accepted",
	)
	klog.Infof("ResourceSlice %q resources condition accepted", client.ObjectKeyFromObject(resourceSlice))

	return ctrl.Result{}, nil
}

// getCappedResources sets the requested resources, but caps the amount of resources
// to the maximum resources defined in the handler.
func (h *Handler) getCappedResources(reqResources corev1.ResourceList) corev1.ResourceList {
	cappedResources := corev1.ResourceList{}

	for name, reqQuantity := range reqResources {
		capQuantity, ok := h.capResources[name]
		if ok && reqQuantity.Cmp(capQuantity) > 0 {
			cappedResources[name] = capQuantity
		} else {
			cappedResources[name] = reqQuantity
		}
	}

	return cappedResources
}
