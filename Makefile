BINARY_NAME=mkcr

.PHONY: build install clean test run

build:
	@go build -o $(BINARY_NAME) .

install: build
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

test:
	@go test ./...

run: build
	./$(BINARY_NAME) $(ARGS)
