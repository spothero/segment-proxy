VERSION_MAJOR ?= local
VERSION_MINOR ?= local
VERSION_PATCH ?= local
VERSION ?= ${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_PATCH}

default_target: all

.PHONY: bootstrap build debug vendor clean clean_vendor test lint docker docker_build docker_run docker_push

all: bootstrap vendor lint build

# Bootstrapping for base golang package deps
BOOTSTRAP=\
	github.com/golang/dep/cmd/dep \
	github.com/alecthomas/gometalinter

$(BOOTSTRAP):
	go get -u $@

bootstrap: $(BOOTSTRAP)
	gometalinter --install

build:
	go build -o bin/segment-proxy -ldflags="-X main.version=${VERSION}" main.go

vendor:
	dep ensure -v -vendor-only

clean:
	rm -rf vendor

test:
	go test -v -coverprofile=coverage.out ./... -cover

coverage: test
	go tool cover -html=coverage.out

# Linting
LINTERS=gofmt golint gosimple vet misspell ineffassign deadcode
METALINT=gometalinter --tests --disable-all --vendor --deadline=5m -e "zz_.*\.go" ./...

lint:
	$(METALINT) $(addprefix --enable=,$(LINTERS))

$(LINTERS):
	$(METALINT) --enable=$@

#################################################
# Docker Commands
#################################################

docker: docker_push

docker_build:
	docker build -t "spothero/segment-proxy:${VERSION}" .

docker_push: docker_build
	docker push spothero/segment-proxy:${VERSION}
