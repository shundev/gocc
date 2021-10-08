int main() {
  char a,b[10],c;

  b[0] = 3;
  b[9] = 4;

  assert(b[0] + b[9], 7);
  return 0;
}
