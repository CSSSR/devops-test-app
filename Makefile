
.PHONY: publish-app start-app deploy

HELM=$(shell which helm3 2>/dev/null || which helm)
BRANCH ?= master

start-app:
	$(MAKE) -C app start

publish-app:
	$(MAKE) -C app build version=$(BRANCH)
	$(MAKE) -C app publish version=$(BRANCH)

deploy:
	$(HELM) upgrade --install my-app-$(BRANCH) chart --set image.tag=$(BRANCH) --set ingress.host=$(BRANCH).csssr.cloud
