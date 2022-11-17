VERSION ?= latest
REGISTRY ?= yndd
IMG ?= $(REGISTRY)/nephio-upf-ipam-fn:${VERSION}

.PHONY: all
all: test

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

test: fmt vet ## Run tests.
	kpt fn render data
	kpt fn render nodata
	kpt fn render dataempty

docker-build: test ## Build docker images.
	docker build -t ${IMG} .

docker-push: ## Build docker images.
	docker push ${IMG}