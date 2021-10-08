int main() {
  int x = 3;
  int y = 5;
  assert( *(&x + 1), 5 );
  return 0;
}
