#!/bin/bash
err="./logs/build.log"

FILES=`find testcases -name '*.c'`
for f in $FILES; do
  echo "$f"
  timeout 3 ./main "$f" 1> tmp.s 2>>$err
  if [[ "$?" != "0" ]]; then
    echo "Error while compiling. Check out $err."
    exit 1
  fi

  cc -o tmp tmp.s hello.o test.o
  ./tmp
  if [[ "$?" != "0" ]]; then
    echo "FAIL"
    exit 1
  fi
done

echo PASS

