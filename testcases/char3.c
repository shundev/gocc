int main() {
  char a = 3;
  char b = 5;
  char c = 9;
  assert( *(&a+2), 9);
  return 0;
}
