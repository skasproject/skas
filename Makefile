
# Image URL to use all building/pushing image targets
DOCKER_IMG := ghcr.io/skasproject/skas

DOCKER_TAG := 0.1.0
VERSION ?= 0.1.0

BUILDX_CACHE=/tmp/docker_cache

# You can switch between simple (faster) docker build or multiplatform one.
# For multiplatform build on a fresh system, do 'make docker-set-multiplatform-builder'
#DOCKER_BUILD := docker buildx build --builder multiplatform --cache-to type=local,dest=$(BUILDX_CACHE),mode=max --cache-from type=local,src=$(BUILDX_CACHE) --platform linux/amd64,linux/arm64
DOCKER_BUILD := docker build

# To authenticate for pushing in github repo:
# echo $GITHUB_TOKEN | docker login ghcr.io -u $USER_NAME --password-stdin

# Comment this to just build locally
DOCKER_PUSH := --push


.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


.PHONY: version
version: ## Set version in binary
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-crd/internal/config/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-ldap/internal/config/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-static/internal/config/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-merge/internal/config/version_.go


.PHONY: docker
docker: version ## Build and push skas image
	$(DOCKER_BUILD) $(DOCKER_PUSH) -t $(DOCKER_IMG):$(DOCKER_TAG) -f Dockerfile .

# ----------------------------------------------------------------------Docker local config

.PHONY: docker-set-multiplatform-builder
docker-set-multiplatform-builder:  ## TO EXECUTE ONCE ON build system, to allow multiplatform build with cache
	docker buildx create --name multiplatform --driver docker-container

