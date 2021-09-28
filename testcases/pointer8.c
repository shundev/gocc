int foo(int *x) {
  *x = 10;
  return 0;
}

int main() {
  int x = 5;
  foo(&x);
  return x;
}
