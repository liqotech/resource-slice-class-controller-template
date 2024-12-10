package handler

import (
	"context"

	authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Handler defines the interface for handling ResourceSlice operations
type Handler interface {
	// Handle processes a ResourceSlice and returns a reconciliation result
	Handle(ctx context.Context, resourceSlice *authv1beta1.ResourceSlice) (ctrl.Result, error)
}
