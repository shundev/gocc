all: test

main: main.go token/*.go ffmt/*.go
	go build main.go

build: main

test: main
	go test ./parser ./token
	./test.sh

repl:
	go run cmd/repl/main.go

clean:
	rm -f main tmp* *.o *~

.PHONY: test build clean repl
