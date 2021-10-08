int main() {
  int x = 3;
  int y = 5;
  assert( *(&y - 1), 3 );
  return 0;
}
