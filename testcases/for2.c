int main() {
  int i = 0;
  int a = 10;
  for (; i < 10;)
    i = i + 1;
  assert( a + i, 20 );
  return 0;
}
