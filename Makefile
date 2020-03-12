all:
	$(MAKE) build
	$(MAKE) check

build:
	go build

check:
	find . -name '*.go' | xargs gofmt -l | xargs echo 'gofmt -w '
	go vet
