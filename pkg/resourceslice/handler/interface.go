// Package handler contains the interface for an handler that manages ResourceSlices resources.
package handler

import (
	"context"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Handler defines the interface for handling ResourceSlice operations.
type Handler interface {
	// Handle processes a ResourceSlice and returns a reconciliation result.
	// The handler should also update the ResourceSlice status. In particular, it should:
	// - set the list of resources that have been allocated in `status.resources`
	// - set a Condition of type "Resources"  in `status.conditions` to indicate
	//   if the resources have been accepted or denied.
	// An error should be returned if the handler fails to process the ResourceSlice
	// and the reconciliation should be retried.
	// Note: it should not return an error if the resourceslice has been correctly processed,
	// but the resources have been denied.
	Handle(ctx context.Context, resourceSlice *authv1beta1.ResourceSlice) (ctrl.Result, error)
}
