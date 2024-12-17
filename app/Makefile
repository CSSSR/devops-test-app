.PHONY: start build publish

export GOBIN=${CURDIR}/bin

version ?= master
registry ?= quay.csssr.cloud

$(GOBIN)/air:
	go install github.com/air-verse/air@latest

start: $(GOBIN)/air
	$(GOBIN)/air

build:
	docker build . --tag $(registry)csssr/test-app:$(version)

publish:
	docker push $(registry)csssr/test-app:$(version)
