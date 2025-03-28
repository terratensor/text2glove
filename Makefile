.PHONY: build test clean

BINARY_NAME = text2glove

build:
	go build -o bin/$(BINARY_NAME) ./cmd/text2glove

test:
	go test ./...

clean:
	rm -rf bin/

install:
	go install ./cmd/text2glove

run:
	go run ./cmd/text2glove -input ./data -output output.txt