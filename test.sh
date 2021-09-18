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

assert 0 "0;"
assert 42 "42;"
assert 35 " 10 + 25 ;"
assert 21 " 5 + 20 - 4;"
assert 30 " 5 + 5 * 5;"
assert 25 " (5 + 5) * 5 / 2;"
assert 2 " 2/1;"
assert 5 "-3 + 8;"
assert 1 "-3 + 8 == 5;"
assert 0 "(5 * 5) == (5 * 2);"
assert 1 "(5 * 5) != (5 * 2);"
assert 0 "5 < 5 == 5 > 2;"
assert 1 "5 <= 5 == 5 >= 2;"
assert 7 "a = 7;"
assert 70 "a = 70; a;"
assert 9 "a = 3; b = 6; a + b;"
assert 18 "c = 9; b = 6; a = 3; a + b + c;"
assert 10 "a = 10; return a; return 20;"
assert 50 "a = 10;b = c= 20; return a + b + c;"
assert 10 "a = 0; if (a + 1) return 10; else return 5;"
assert 5 "a = 0; if (a) return 10; else return 5;"
assert 8 "a = 0; while (a < 5) a = a + 4; return a;"

echo OK
