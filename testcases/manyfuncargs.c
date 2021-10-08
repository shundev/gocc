int f(int a, int b, int c, int d, int e, int g, int h, int i, int j) {
  int k = 1, l = 1;
  return a + b + c + d + e + g + h + i + j + k + l;
}

int main() {
  assert( f(1, 1, 1, 1, 1, 1, 1, 1, 1), 11 );
  return 0;
}
