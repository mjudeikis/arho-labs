TAG=$(shell git describe --tags HEAD)$(shell [[ $$(git status --porcelain) = "" ]] && echo -clean || echo -dirty)
LDFLAGS="-X main.gitCommit=$(TAG)"
IMAGEBUILDER = ${GOPATH}/bin/imagebuilder

image-builder:
	go get -u github.com/openshift/imagebuilder/cmd/imagebuilder

frontend:
	go build -ldflags ${LDFLAGS} ./cmd/$@

image: image-builder frontend
	$(IMAGEBUILDER) -f Dockerfile -t quay.io/mangirdas/osa-labs .

run: 
	go run -ldflags ${LDFLAGS} ./cmd/app -dev-mode

push: image
	docker push quay.io/mangirdas/osa-labs
