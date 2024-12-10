# Resource Slice Classes Controller

A Kubernetes controller for managing ResourceSlice resources with customizable behavior based on class names. This controller allows you to implement different resource allocation strategies for different classes of ResourceSlices.

## Overview

The Resource Slice Classes controller is designed to manage ResourceSlice resources in a Kubernetes cluster. Each controller instance handles ResourceSlices of a specific class, allowing you to run multiple controllers with different behaviors for different classes.

The controller manages the ResourceSlice status updates and conditions, while the handler is responsible for implementing the resource allocation strategy.

## Features

- Class-based resource handling
- Pluggable handler interface for custom resource allocation strategies
- Automatic status and condition management
- Example implementation included
- Built on controller-runtime

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/liqotech/resource-slice-classes.git
   cd resource-slice-classes
   ```

2. Build the controller:
   ```bash
   go build -o bin/manager main.go
   ```

## Usage

### Running the Controller

The controller requires a class name to be specified:

```bash
./bin/manager --class-name=my-class
```

Additional flags:
- `--metrics-bind-address`: The address to bind the metrics endpoint (default: ":8080")
- `--health-probe-bind-address`: The address to bind the health probe endpoint (default: ":8081")
- `--leader-elect`: Enable leader election for controller manager

### Example Implementation

The repository includes an example handler implementation in `example/resourceslice/handler.go` that:
- Generates CPU resources between 1 and 10 cores
- Generates Memory resources between 1 and 5 GB
- Allocates 110 pods
- Uses a deterministic hash of the ResourceSlice name for consistent resource allocation

## Creating Custom Handlers

To implement a custom handler for your ResourceSlice class:

1. Create a new type that implements the `handler.Handler` interface:

```go
package myhandler

import (
    "context"
    authv1beta1 "github.com/liqotech/liqo/apis/authentication/v1beta1"
    "github.com/liqotech/resource-slice-classes/pkg/resourceslice/handler"
    ctrl "sigs.k8s.io/controller-runtime"
)

type MyHandler struct {}

func NewMyHandler() handler.Handler {
    return &MyHandler{}
}

func (h *MyHandler) Handle(ctx context.Context, resourceSlice *authv1beta1.ResourceSlice) (ctrl.Result, error) {
    // Implement your custom resource allocation logic here
    // Update resourceSlice.Status.Resources with your allocated resources
    
    return ctrl.Result{}, nil
}
```

2. Update `main.go` to use your custom handler:

```go
import (
    "github.com/your-org/your-module/pkg/myhandler"
)

func main() {
    // ...
    
    // Create your custom handler
    customHandler := myhandler.NewMyHandler()
    
    if err = controller.NewResourceSliceReconciler(
        mgr.GetClient(),
        mgr.GetScheme(),
        mgr.GetEventRecorderFor("resource-slice-controller"),
        className,
        customHandler,
    ).SetupWithManager(mgr); err != nil {
        // ...
    }
    
    // ...
}
```

## Handler Interface

The handler interface is defined in `pkg/resourceslice/handler/interface.go`:

```go
type Handler interface {
    Handle(ctx context.Context, resourceSlice *authv1beta1.ResourceSlice) (ctrl.Result, error)
}

Your handler implementation should:
1. Implement your resource allocation strategy
2. Set the allocated resources in `resourceSlice.Status.Resources`
3. Return appropriate reconciliation results and errors

Note: The controller, not the handler, is responsible for:
- Updating the ResourceSlice status in the API server
- Managing ResourceSlice conditions
- Recording events
- Error handling and logging

## Best Practices

1. **Resource Allocation**:
   - Make your allocation strategy deterministic when possible
   - Consider resource constraints and quotas
   - Focus on the allocation logic, leaving status updates to the controller

2. **Handler Implementation**:
   - Keep handlers focused on resource calculation logic
   - Return meaningful errors for proper event recording
   - Use logging for debugging purposes

3. **Error Handling**:
   - Return errors when resource allocation fails
   - Let the controller handle retries and status updates
   - Use appropriate error types for different failure scenarios

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
