#!/bin/bash
assert() {
  expected="$1"
  input="$2"
  err="./logs/build.log"

  timeout 3 ./main "testcases/$input" 1> tmp.s 2>>$err
  if [[ "$?" != "0" ]]; then
    echo "Error while compiling. Check out $err."
    exit 1
  fi

  cc -o tmp tmp.s hello.o
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assert 15 global2.c
assert 0 global1.c
assert 2 array1.c
assert 10 array2.c
assert 6  func1.c
assert 8  func2.c
#assert 11 manyfuncargs.c
assert 55 fib.c
assert 14 sizeof1.c
assert 18 sizeof2.c
assert 4  sizeof3.c
assert 8  sizeof4.c
assert 4  sizeof5.c
assert 22 assign.c
assert 35 ident.c
assert 3  pointer1.c
assert 7  pointer2.c
assert 5  pointer3.c
assert 3  pointer4.c
assert 5  pointer5.c
assert 5  pointer6.c
assert 10 pointer7.c
assert 10 pointer8.c
assert 30 arith1.c
assert 25 arith2.c
assert 0  comp1.c
assert 1  comp2.c
assert 0  comp3.c
assert 1  comp4.c
assert 10 return.c
assert 10 if1.c
assert 15 if2.c
assert 8  while1.c
assert 12 for1.c
assert 20 for2.c
echo OK
