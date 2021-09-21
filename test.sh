#!/bin/bash
assert() {
  expected="$1"
  input="$2"

  timeout 3 ./main "$input" 1> tmp.s 2>>./logs/build.log
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

assert 3 "x=3; y=5; z=7; return *(&x+0);"
assert 7 "x=3; y=5; z=7; return *(&x+2);"
assert 5 "x=3; y=5; return *(&x+1);"
assert 3 "x=3; y=5; return *(&y-1);"
assert 5 "x = 5;return *&*&x;"
assert 5 "x = 5; y = &x; z = &y; return **z;"
assert 5 "x = 5;return *&x;"
assert 2 "bar = 2; return bar;"
assert 1 "bar = 1; return bar;"
assert 35 "abc123 = 5; 🍺🍣 = 10; ホゲホゲ=20; return abc123 + 🍺🍣 + ホゲホゲ;"
assert 22 "a=b=c=d=e=f=g=h=i=j=k=2; return a+b+c+d+e+f+g+h+i+j+k;"
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
assert 12 "for (i = 0; i < 10; i = i + 4) 4; return i;"
assert 20 "i = 0; a = 10; for (; i < 10; ) i = i + 1; return a + i;"
assert 15 "if (1) { a = 5; b = 10; 20; return a + b; 30; } else return 50;"

echo OK
