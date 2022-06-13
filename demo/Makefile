.PHONY: build
build: pre-build
	@go build -o target/demo main.go

.PHONY: pre-build
pre-build:
	@mkdir -p target

.PHONY: controller-gen
controller-gen:
	@controller-gen object paths=./api/...

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: lint
lint:
	@go vet ./...

.PHONY: clean
clean:
	@rm -r target