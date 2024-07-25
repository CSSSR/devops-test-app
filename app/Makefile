.PHONY: start build publish

export GOBIN=${CURDIR}/bin

version ?= master

$(GOBIN)/air:
	go install github.com/air-verse/air@latest

start: $(GOBIN)/air
	$(GOBIN)/air

build:
	docker build . --tag quay.csssr.cloud/csssr/test-app:$(version)

publish:
	docker push quay.csssr.cloud/csssr/test-app:$(version)
