int x = 5;
int *y = &x;

int main() {
  *y = 8;
  assert( x, 8 );
  return 0;
}
