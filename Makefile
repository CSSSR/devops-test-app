
.PHONY: publish-app start-app deploy

HELM=$(shell which helm3 2>/dev/null || which helm)
NAMESPACE ?= csssr-test-app
BRANCH ?= master
DOCKER_REGISTRY ?= quay.csssr.cloud
PULL_POLICY ?= Always
ROOT_DOMAIN_NAME ?= my-app.com

start-app:
	$(MAKE) -C app start

publish-app:
	$(MAKE) -C app build version=$(BRANCH) registry=$(DOCKER_REGISTRY)
	$(MAKE) -C app publish version=$(BRANCH) registry=$(DOCKER_REGISTRY)

deploy:
	$(HELM) dependency list chart
	$(HELM) dependency update chart
	$(HELM) upgrade \
	    --install my-app-$(BRANCH) chart \
	    --namespace $(NAMESPACE) \
	    --set image.name=$(DOCKER_REGISTRY)/csssr/test-app \
	    --set image.tag=$(BRANCH) \
	    --set image.pullPolicy=$(PULL_POLICY) \
	    --set ingress.host=$(BRANCH)-cssr-devops-test-app.$(ROOT_DOMAIN_NAME) \
	    --set ingress.tls.enabled=true \
	    --set deployment.replicas=$(REPLICAS)

undeploy:
	$(HELM) uninstall ${APP_TAG} --namespace ${NAMESPACE}