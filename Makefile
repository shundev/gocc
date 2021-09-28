all: test

main: main.go token/*.go parser/*.go generator/*.go repl/*.go
	go build main.go

build: main

test: main
	go test ./parser ./token ./generator ./repl
	./test.sh

repl:
	go run cmd/repl/main.go

clean:
	rm -f main tmp* *.o *~

.PHONY: test build clean repl
