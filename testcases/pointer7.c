int main() {
  int x = 5;
  int y = &x;
  *y = 10;
  return x;
}
