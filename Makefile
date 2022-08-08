all: build docker

build:
	@GOBIN=`pwd` CGO_ENABLED=0 go install --ldflags '-extldflags "-static"'

docker:
	@docker build -t vish/stress .

image-build: build
	docker build -t gcr.io/saikatroyc-test/stress:v1 .

image-push: image-build
	docker push gcr.io/saikatroyc-test/stress:v1

.PHONY: docker build all
