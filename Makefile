TAG=$(shell git describe --tags HEAD)$(shell [[ $$(git status --porcelain) = "" ]] && echo -clean || echo -dirty)
LDFLAGS="-X main.gitCommit=$(TAG)"
IMAGEBUILDER = ${GOPATH}/bin/imagebuilder

dep:
	export GO111MODULE=on
	gvm use go1.12.2
	go mod tidy

image-builder:
	go get -u github.com/openshift/imagebuilder/cmd/imagebuilder

frontend:
	go build -ldflags ${LDFLAGS} ./cmd/$@

ssh:
	go build -ldflags ${LDFLAGS} ./cmd/$@

image: image-builder frontend ssh
	$(IMAGEBUILDER) -f Dockerfile -t quay.io/mangirdas/labs-frontend .
	$(IMAGEBUILDER) -f Dockerfile.worker -t quay.io/mangirdas/labs-worker  .

run: 
	go run -ldflags ${LDFLAGS} ./cmd/app -dev-mode

push: image
	docker push quay.io/mangirdas/labs-frontend
	docker push quay.io/mangirdas/labs-ssh


setup-cluster:
	oc new-project summit
	oc new-project workers
