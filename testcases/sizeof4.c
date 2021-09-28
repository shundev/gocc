int *foo() {
  int x = 10;
  return &x;
}

int main() {
  return sizeof foo();
}
