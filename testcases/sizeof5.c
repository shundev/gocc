int foo() {
  int x = 10;
  return x;
}

int main() {
  assert( sizeof foo(), 4 );
  return 0;
}
