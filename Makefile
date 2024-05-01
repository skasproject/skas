
# Image URL to use all building/pushing image targets
DOCKER_IMG := ghcr.io/skasproject/skas

VERSION := 0.2.2-snapshot
DOCKER_TAG := 0.2.2-snapshot
# Must be in sync with corresponding Chart.yaml
CHART_VERSION ?= 0.2.2-snapshot
SK_USERS_CHART_VERSION ?= 0.2.2

BUILD_TS ?= $(shell date -u +%Y%m%d.%H%M%S)

SKAS_CHARTS ?= "../skas-charts"
#SKAS_CHARTS_PRE_ALPHA ?= "../skas-charts-pre-alpha"

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
	@echo "// Generated by Makefile\n\npackage config\n\nvar Version = \"$(VERSION)\"\nvar BuildTs = \"$(BUILD_TS)\"\n" >sk-common/pkg/config/version_.go

.PHONY: docker
docker: version ## Build and push skas image (And update version)
	$(DOCKER_BUILD) $(DOCKER_PUSH) -t $(DOCKER_IMG):$(DOCKER_TAG) -f Dockerfile .

.PHONY: roles
roles: ## Publish ansible roles in a public repo
	cd extra/ansible/roles && tar cvzf ../../dist/skas-apiserver-role-${VERSION}.tgz skas-apiserver/
	cd ../warehouse && gh release upload --clobber $(VERSION) ../skas/extra/dist/skas-apiserver-role-${VERSION}.tgz

#.PHONY: charts
#charts: ## Publish helm chart in a public repo (Not the main one)
#	cd extra/helm && helm package -d ../dist skas
#	cd ../warehouse && gh release upload  --clobber $(VERSION) ../skas/extra/dist/skas-$(VERSION).tgz
#	cd extra/helm && helm package -d ../dist skusers
#	cd ../warehouse && gh release upload  --clobber $(VERSION) ../skas/extra/dist/skusers-$(VERSION).tgz

.PHONY: manifests
manifests:	## Generate CustomResourceDefinition manifests.
	cd sk-auth && make manifests
	cd sk-common && make manifests

.PHONY: generate
generate:	## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	cd sk-auth && make generate
	cd sk-common && make generate

.PHONY: doc
doc: ## Generate doc index
	doctoc docs/installation.md --github --title '## Index'
	doctoc docs/usage.md --github --title '## Index'

#.PHONY: charts-pre-alpha
#charts-pre-alpha: ## Publish helm chart in a private (pre-alpha) repo
#	cd extra/helm && helm package -d ../dist skas
#	cp extra/dist/skas-$(VERSION).tgz $(SKAS_CHARTS_PRE_ALPHA)/charts/skas-$(VERSION).tgz
#	cd $(SKAS_CHARTS_PRE_ALPHA) && helm repo index --url https://skasproject.github.io/skas-charts-pre-alpha . && git add . && git commit -m "Update charts" && git push


.PHONY: charts
charts: ## Publish helm chart in our public repo
	cd extra/helm && helm package -d ../dist skas
	cd extra/helm && helm package -d ../dist skusers
	cd $(SKAS_CHARTS) && git pull
	cp extra/dist/skas-$(CHART_VERSION).tgz $(SKAS_CHARTS)/charts/skas-$(CHART_VERSION).tgz
	cp extra/dist/skusers-$(SK_USERS_CHART_VERSION).tgz $(SKAS_CHARTS)/charts/skusers-$(SK_USERS_CHART_VERSION).tgz
	cd $(SKAS_CHARTS) && helm repo index --url https://skasproject.github.io/skas-charts . && git add . && git commit -m "Update charts" && git push

# .PHONY: charts
# charts: ## Publish helm chart in a public repo
# 	cd extra/helm && helm package -d ../dist skas
# 	cp extra/dist/skas-$(VERSION).tgz $(SKAS_CHARTS)/charts/skas-$(VERSION).tgz
# 	cd $(SKAS_CHARTS) && helm repo index --url https://skasproject.github.io/skas-charts . && git add . && git commit -m "Update charts" && git push
#

# Commented out to avoid mistake. Remove comment to publish
# .PHONY: docsite
# docsite: ## Publish documentation
# 	cd docsite && . ./setup/activate.sh && mkdocs gh-deploy --clean --force

# ----------------------------------------------------------------------Docker local config

.PHONY: docker-set-multiplatform-builder
docker-set-multiplatform-builder:  ## TO EXECUTE ONCE ON build system, to allow multiplatform build with cache
	docker buildx create --name multiplatform --driver docker-container

