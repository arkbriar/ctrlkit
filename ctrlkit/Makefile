.PHONY: build
build: pre-build
	@go build -o target/ctrlkit cmd/ctrlkit.go 

.PHONY: pre-build
pre-build:
	@mkdir -p target

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: lint
lint:
	@go vet ./...

.PHONY: clean
clean:
	@rm -r target