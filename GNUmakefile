default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...
	git add docs/
	git commit -m "Update generated documentation" || echo "No changes to commit."

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...


.PHONY: fmt lint test testacc build install generate
