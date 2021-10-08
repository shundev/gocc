all: test

main: main.go token/*.go parser/*.go generator/*.go repl/*.go writer/*.go types/*.go ast/*.go
	go build main.go

build: main

hello.o: c/hello.c
	cc -c c/hello.c

test.o: c/test.c
	cc -c c/test.c

test: main hello.o test.o
	go test ./parser ./token ./generator ./repl ./writer
	./test.sh

repl:
	go run cmd/repl/main.go

clean:
	rm -f main tmp* *.o *~

# エラーになるかもしれないが、tmpのステータスコードが表示される
sample: main hello.o
	./main testcases/sample.c >./tmp.s 2>>./logs/build.log
	cc -o tmp tmp.s
	./tmp

asm:
	cc -o tmp tmp.s
	./tmp

.PHONY: test build clean repl sample asm
