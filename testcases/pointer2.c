int main() {
  int x = 3;
  int y = 5;
  int z = 7;
  assert( *(&x + 2), 7 );
  return 0;
}
