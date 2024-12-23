// Package controller contains the ResourceSlice reconciler.
package controller

import (
	"context"
	"fmt"
	"strconv"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	"github.com/liqotech/liqo/pkg/consts"
	"github.com/liqotech/liqo/pkg/liqo-controller-manager/authentication"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	rshandler "github.com/liqotech/resource-slice-class-controller-template/pkg/resourceslice/handler"
)

// ResourceSliceReconciler reconciles a ResourceSlice object.
type ResourceSliceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	recorder  record.EventRecorder
	className string
	handler   rshandler.Handler
}

// NewResourceSliceReconciler creates a new ResourceSliceReconciler.
func NewResourceSliceReconciler(cl client.Client, scheme *runtime.Scheme, recorder record.EventRecorder,
	className string, handler rshandler.Handler) *ResourceSliceReconciler {
	return &ResourceSliceReconciler{
		Client:    cl,
		Scheme:    scheme,
		recorder:  recorder,
		className: className,
		handler:   handler,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceSliceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// generate the predicate to filter just the ResourceSlices that are replicated
	// (i.e., the ones for which we have the role of provider)
	replicatedResSliceFilter, err := predicate.LabelSelectorPredicate(replicatedResourcesLabelSelector())
	if err != nil {
		klog.Error(err)
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1beta1.ResourceSlice{}, builder.WithPredicates(replicatedResSliceFilter)).
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
	klog.V(4).Infof("Reconciling ResourceSlice %q (class: %q)", req.NamespacedName, r.className)

	// Fetch the ResourceSlice instance
	var resourceSlice authv1beta1.ResourceSlice
	if err = r.Get(ctx, req.NamespacedName, &resourceSlice); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Wait for the Authentication condition to be ready
	authCond := authentication.GetCondition(&resourceSlice, authv1beta1.ResourceSliceConditionTypeAuthentication)
	if authCond == nil || authCond.Status != authv1beta1.ResourceSliceConditionAccepted {
		return ctrl.Result{}, nil
	}

	// Delegate the handling to the handler
	res, err = r.handler.Handle(ctx, &resourceSlice)
	if err != nil {
		r.recorder.Eventf(&resourceSlice, "Warning", "Failed", "Failed to handle ResourceSlice: %v", err)
		return ctrl.Result{}, fmt.Errorf("failed to handle ResourceSlice %q: %w", req.NamespacedName, err)
	}

	// check the "Resources" condition is set.
	resCond := authentication.GetCondition(&resourceSlice, authv1beta1.ResourceSliceConditionTypeResources)
	if resCond == nil {
		return ctrl.Result{}, fmt.Errorf("failed to handle ResourceSlice %q: missing \"Resources\" condition", req.NamespacedName)
	}

	defer func() {
		// Update the status
		if newErr := r.Status().Update(ctx, &resourceSlice); newErr != nil {
			if err != nil {
				klog.Error(err)
			}
			r.recorder.Eventf(&resourceSlice, "Warning", "Failed", "Failed to update ResourceSlice status: %v", err)
			err = fmt.Errorf("failed to update ResourceSlice %q status: %w", req.NamespacedName, newErr)
			return
		}
		klog.Infof("ResourceSlice %q status correctly updated", req.NamespacedName)
	}()

	// Return the reconciliation result
	return res, nil
}

// replicatedResourcesLabelSelector is an helper function which returns a label selector to list all the replicated resources.
func replicatedResourcesLabelSelector() metav1.LabelSelector {
	return metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      consts.ReplicationOriginLabel,
				Operator: metav1.LabelSelectorOpExists,
			},
			{
				Key:      consts.ReplicationStatusLabel,
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{strconv.FormatBool(true)},
			},
		},
	}
}
