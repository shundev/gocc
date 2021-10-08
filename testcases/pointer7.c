int main() {
  int x = 5;
  int *y = &x;
  *y = 10;
  assert( x, 10 );
  return 0;
}
