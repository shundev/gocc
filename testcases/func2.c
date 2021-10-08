int f(int a, int b, int c, int d, int e, int g) {
  int h = 1, i = 1;
  return a + b + c + d + e + g + h + i;
}

int main() {
  assert( f(1, 1, 1, 1, 1, 1), 8 );
  return 0;
}
