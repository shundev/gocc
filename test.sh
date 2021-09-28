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

assert 33 "int main () { return add(10, 23); }"
assert 13 "int hoge () { int a = 3; return a; } int fuga() { int a = 10; return a; } int main() { int a = hoge(); int b = fuga(); return a + b; }"
assert 35 "int hoge () { return 10; } int fuga() { return 25; } int main() { return hoge() + fuga(); }"
assert 22 "int main () { int a,b,c,d,e,f,g,h,i,j,k; a=b=c=d=e=f=g=h=i=j=k=2; return a+b+c+d+e+f+g+h+i+j+k; }"
assert 3  "int main () { int x=3; int y=5; int z=7; return *(&x+0); }"
assert 7  "int main () { int x=3; int y=5; int z=7; return *(&x+2) ; }"
assert 5  "int main () { int x=3; int y=5; return *(&x+1); }"
assert 3  "int main () { int x=3; int y=5; return *(&y-1); }"
assert 5  "int main () { int x = 5;return *&*&x; }"
assert 5  "int main () { int x = 5; int *y = &x; int **z = &y; return **z; }"
assert 5  "int main () { int x = 5;return *&x; }"
assert 2  "int main () { int bar = 2; return bar; }"
assert 1  "int main () { int bar = 1; return bar; }"
assert 35 "int main () { int abc123 = 5; int 🍺🍣 = 10; int ホゲホゲ=20; return abc123 + 🍺🍣 + ホゲホゲ; }"
assert 0  "int main () { 0; }"
assert 42 "int main () { 42; }"
assert 35 "int main () {  10 + 25 ; }"
assert 21 "int main () {  5 + 20 - 4; }"
assert 30 "int main () {  5 + 5 * 5; }"
assert 25 "int main () {  (5 + 5) * 5 / 2; }"
assert 2  "int main () {  2/1; }"
assert 5  "int main () { -3 + 8; }"
assert 1  "int main () { -3 + 8 == 5; }"
assert 0  "int main () { (5 * 5) == (5 * 2); }"
assert 1  "int main () { (5 * 5) != (5 * 2); }"
assert 0  "int main () { 5 < 5 == 5 > 2; }"
assert 1  "int main () { 5 <= 5 == 5 >= 2; }"
assert 7  "int main () { int a = 7; }"
assert 70 "int main () { int a = 70; a; }"
assert 9  "int main () { int a = 3; int b = 6; a + b; }"
assert 18 "int main () { int c = 9; int b = 6; int a = 3; a + b + c; }"
assert 10 "int main () { int a = 10; return a; return 20; }"
assert 10 "int main () { int a = 0; if (a + 1) return 10; else return 5; }"
assert 5  "int main () { int a = 0; if (a) return 10; else return 5; }"
assert 8  "int main () { int a = 0; while (a < 5) a = a + 4; return a; }"
assert 12 "int main () { for (int i = 0; i < 10; i = i + 4) 4; return i; }"
assert 20 "int main () { int i = 0; int a = 10; for (; i < 10; ) i = i + 1; return a + i; }"
assert 15 "int main () { if (1) { int a = 5; int b = 10; 20; return a + b; 30; } else return 50; }"
echo OK
