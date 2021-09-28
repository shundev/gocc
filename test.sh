#!/bin/bash
assert() {
  expected="$1"
  input="$2"

  timeout 3 ./main "testcases/$input" 1> tmp.s 2>>./logs/build.log
  cc -o tmp tmp.s
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assert 66 manyfuncargs.c
assert 55 fib.c
assert 22 assign.c
assert 35 ident.c
assert 3  pointer1.c
assert 7  pointer2.c
assert 5  pointer3.c
assert 3  pointer4.c
assert 5  pointer5.c
assert 5  pointer6.c
assert 30 arith1.c
assert 25 arith2.c
assert 0  comp1.c
assert 1  comp2.c
assert 0  comp3.c
assert 1  comp4.c
assert 10 return.c
assert 10 if1.c
assert 15 if2.c
assert 8 while1.c
assert 12 for1.c
assert 20 for2.c
echo OK
