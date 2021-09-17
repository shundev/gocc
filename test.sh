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
assert 30 " 5 + 5 * 5"
assert 25 " (5 + 5) * 5 / 2"
assert 2 " 2/1"
assert 5 "-3 + 8"
assert 1 "-3 + 8 == 5"
assert 0 "(5 * 5) == (5 * 2)"
assert 1 "(5 * 5) != (5 * 2)"
assert 0 "5 < 5 == 5 > 2"
assert 1 "5 <= 5 == 5 >= 2"

echo OK
