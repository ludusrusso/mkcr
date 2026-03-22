BINARY_NAME=mkcr

.PHONY: build install clean test run lint

build:
	@go build -o $(BINARY_NAME) .

install: build
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

test:
	@go test ./...

lint:
	@golangci-lint run ./...

run: build
	./$(BINARY_NAME) $(ARGS)
