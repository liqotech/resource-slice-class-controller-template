package controller

import (
	"context"
	"fmt"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	"github.com/liqotech/liqo/pkg/liqo-controller-manager/authentication"
	"github.com/liqotech/resource-slice-classes/pkg/resourceslice/handler"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ResourceSliceReconciler reconciles a ResourceSlice object.
type ResourceSliceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	recorder  record.EventRecorder
	className string
	handler   handler.Handler
}

// NewResourceSliceReconciler creates a new ResourceSliceReconciler.
func NewResourceSliceReconciler(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder, className string, handler handler.Handler) *ResourceSliceReconciler {
	return &ResourceSliceReconciler{
		Client:    client,
		Scheme:    scheme,
		recorder:  recorder,
		className: className,
		handler:   handler,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceSliceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1beta1.ResourceSlice{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			resourceSlice, ok := obj.(*authv1beta1.ResourceSlice)
			if !ok {
				return false
			}
			return string(resourceSlice.Spec.Class) == r.className
		})).
		Complete(r)
}

// Reconcile handles the reconciliation loop for ResourceSlice resources.
func (r *ResourceSliceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	klog.V(4).InfoS("Reconciling ResourceSlice", "name", req.Name, "namespace", req.Namespace, "className", r.className)

	// Fetch the ResourceSlice instance
	resourceSlice := authv1beta1.ResourceSlice{}
	if err = r.Get(ctx, req.NamespacedName, &resourceSlice); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Wait for the Authentication condition to be ready
	if resourceSlice.Status.Conditions == nil {
		return ctrl.Result{}, nil
	}
	if authentication.GetCondition(&resourceSlice, authv1beta1.ResourceSliceConditionTypeAuthentication).Status != authv1beta1.ResourceSliceConditionAccepted {
		return ctrl.Result{}, nil
	}

	// Delegate the handling to the handler
	res, err = r.handler.Handle(ctx, &resourceSlice)
	if err != nil {
		r.recorder.Eventf(&resourceSlice, "Warning", "Failed", "Failed to handle ResourceSlice: %v", err)
		return ctrl.Result{}, err
	}

	defer func() {
		// Update the status
		if err = r.Status().Update(ctx, &resourceSlice); err != nil {
			r.recorder.Eventf(&resourceSlice, "Warning", "Failed", "Failed to update ResourceSlice status: %v", err)
			err = fmt.Errorf("failed to update ResourceSlice status: %w", err)
		}
	}()

	// Update the conditions
	if resourceSlice.Status.Conditions == nil {
		resourceSlice.Status.Conditions = []authv1beta1.ResourceSliceCondition{}
	}
	authentication.EnsureCondition(
		&resourceSlice,
		authv1beta1.ResourceSliceConditionTypeResources,
		authv1beta1.ResourceSliceConditionAccepted,
		"ResourceSliceResourcesAccepted",
		"ResourceSlice resources accepted",
	)
	// Return the reconciliation result
	return res, nil
}
