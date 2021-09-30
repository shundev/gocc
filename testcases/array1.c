int main() {
  int x,y[4],z;
  *y = 1;
  *(y+3) = 4;
  *(y+2) = 3;
  *(y+1) = 2;
  return *y + *(y+1) + *(y+2) + *(y+3);
}
