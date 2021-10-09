#include "stdio.h"
#include "stdlib.h"

int assert(int got, int want) {
  if (want != got) {
    printf("want=%d, but got=%d\n", want, got);
    exit(1);
  }

  return 0;
}

int assertC(char got, char want) {
  if (want != got) {
    printf("want=%c, but got=%c\n", want, got);
    exit(1);
  }

  return 0;
}

int assertS(char* got, char* want, int len) {
  for (int i=0; i<len; i++) {
    if (want[i] != got[i]) {
      printf("want=%s, but got=%s\n", want, got);
      exit(1);
    }
  }

  return 0;
}
