# Run go fmt against code
fmt: gci addlicense
	go mod tidy
	go fmt ./...
	find . -type f -name '*.go' -a ! -name '*zz_generated*' -exec $(GCI) write -s standard -s default -s "prefix(github.com/liqotech/resource-slice-class-controller-template)" {} \;
	find . -type f -name '*.go' -exec $(ADDLICENSE) -l apache -c "The Liqo Authors" -y "2024-$(shell date +%Y)" {} \;
	find . -type f -name "*.go" -exec sed -i "s/Copyright 2024-[0-9]\{4\} The Liqo Authors/Copyright 2024-$(shell date +%Y) The Liqo Authors/" {} +

# Install gci if not available
gci:
ifeq (, $(shell which gci))
	@go install github.com/daixiang0/gci@v0.13.4
GCI=$(GOBIN)/gci
else
GCI=$(shell which gci)
endif

# Install addlicense if not available
addlicense:
ifeq (, $(shell which addlicense))
	@go install github.com/google/addlicense@v1.0.0
ADDLICENSE=$(GOBIN)/addlicense
else
ADDLICENSE=$(shell which addlicense)
endif
