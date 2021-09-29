.intel_syntax noprefix
.globl main
main:
  push rbp
  mov rbp, rsp
  sub rsp, 48
  lea rax, [rbp-48]
  push rax
  mov rax, 1
  pop rdi
  mov [rdi], rax
  lea rax, [rbp-40]
  mov rax, [rax]
  push rax
  mov rax, 2
  pop rdi
  mov [rdi], rax
  lea rax, [rbp-8]
  push rax
  mov rax, 3
  pop rdi
  mov [rdi], rax
  mov rax, 0
  jmp .L.return.main
.L.return.main:
  mov rsp, rbp
  pop rbp
  ret
