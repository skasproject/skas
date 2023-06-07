
# Image URL to use all building/pushing image targets
DOCKER_IMG := ghcr.io/skasproject/skas

DOCKER_TAG := 0.2.1
VERSION ?= 0.2.1

# To authenticate for pushing in github repo:
# echo $GITHUB_TOKEN | docker login ghcr.io -u $USER_NAME --password-stdin

# To authenticate for using gh commands
# gh auth login
# ? What account do you want to log into? GitHub.com
# ? What is your preferred protocol for Git operations? HTTPS
# ? Authenticate Git with your GitHub credentials? Yes
# ? How would you like to authenticate GitHub CLI? Login with a web browser

BUILDX_CACHE=/tmp/docker_cache

# You can switch between simple (faster) docker build or multiplatform one.
# For multiplatform build on a fresh system, do 'make docker-set-multiplatform-builder'
#DOCKER_BUILD := docker buildx build --builder multiplatform --cache-to type=local,dest=$(BUILDX_CACHE),mode=max --cache-from type=local,src=$(BUILDX_CACHE) --platform linux/amd64,linux/arm64
DOCKER_BUILD := docker build


# Comment this to just build locally
DOCKER_PUSH := --push


.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)



.PHONY: all
all: version manifests generate charts		## All but docker

.PHONY: version
version: ## Set version in binary
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-crd/internal/config/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-ldap/internal/config/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-static/internal/config/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-merge/internal/config/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-auth/internal/config/version_.go
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\n" >sk-clientgo/internal/config/version_.go


.PHONY: docker
docker: version ## Build and push skas image
	$(DOCKER_BUILD) $(DOCKER_PUSH) -t $(DOCKER_IMG):$(DOCKER_TAG) -f Dockerfile .


.PHONY: charts
charts: ## Publish helm chart
	cd helm && helm package -d ../dist skas
	cd ../warehouse && gh release upload  --clobber $(VERSION) ../skas/dist/skas-$(VERSION).tgz
	cd helm && helm package -d ../dist skusers
	cd ../warehouse && gh release upload  --clobber $(VERSION) ../skas/dist/skusers-$(VERSION).tgz


.PHONY: manifests
manifests:	## Generate CustomResourceDefinition manifests.
	cd sk-auth && make manifests
	cd sk-common && make manifests

.PHONY: generate
generate:	## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	cd sk-auth && make generate
	cd sk-common && make generate

# ----------------------------------------------------------------------Docker local config

.PHONY: docker-set-multiplatform-builder
docker-set-multiplatform-builder:  ## TO EXECUTE ONCE ON build system, to allow multiplatform build with cache
	docker buildx create --name multiplatform --driver docker-container

