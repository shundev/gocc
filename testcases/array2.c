int main() {
  int x,y[4],z;
  z = 20;
  y[3] = 1;
  y[2] = 2;
  y[1] = 3;
  y[0] = 4;
  x = 10;
  assert(y[0] + y[1] + y[2] + y[3], 10);
  return 0;
}
