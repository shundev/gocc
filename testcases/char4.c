char z(char a, char b, char c, char d, char e) {
  char f = 1;
  return a + b + c + d + e + f;
}

int main() {
  char a = 10;
  char b = 20;
  char c = z(a, b, 1, 1, 1);
  return c;
}
