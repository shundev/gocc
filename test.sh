#!/bin/bash
assert() {
  expected="$1"
  input="$2"

  timeout 3 ./main "$input" 1> tmp.s 2>>./logs/build.log
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

assert 0 0
assert 42 42
assert 35 " 10 + 25 "
assert 21 " 5 + 20 - 4"

echo OK
