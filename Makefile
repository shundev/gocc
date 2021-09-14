all: test

main: main.go
	go build main.go

build: main

test: main
	./test.sh

clean:
	rm -f main tmp* *.o *~

.PHONY: test build clean
