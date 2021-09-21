*a

walk(*a)
  walk(a)
    address(a)
      lea RAX [RBP-offset(a)]
    mov RAX [RAX]
  mov RAX [RAX]

&a
walk(&a)
  address(a)
    lea RAX [RBP-offset(a)]

