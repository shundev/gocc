#include "stdio.h"

int assert(int got, int want) {
  if (want != got) {
    printf("want=%d, but got=%d\n", want, got);
    exit(1);
  }

  return 0;
}
