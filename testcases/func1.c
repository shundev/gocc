int f(int a, int b, int c) {
  return a + b + c;
}

int main() {
  assert(f(1, 2, 3), 6);
  return 0;
}
