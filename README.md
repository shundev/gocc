
# 関数への引数渡し

* 0: RAX
* 1: RSI
* 2: RDX
* 3: RCX
* 4: R8D
* 5: R9D
* 6~: push


```
.intel_syntax noprefix
.globl main
.LC0:
        .string "%d %d %d %d %d %d %d"
main:
        push    rbp
        mov     rbp, rsp
        push    7
        push    6
        mov     r9d, 5
        mov     r8d, 4
        mov     rcx, 3
        mov     rdx, 2
        mov     rsi, 1
        lea     rax, .LC0[rip]
        mov     rdi, rax
        mov     rax, 0
        call    printf
        add     rsp, 16
        mov     rax, 0
        leave
        ret
```
