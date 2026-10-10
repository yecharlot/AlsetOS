/* Multiboot header + entry — AlsetOS Kernel native (no host OS) */
.set ALIGN,    1<<0
.set MEMINFO,  1<<1
.set FLAGS,    ALIGN | MEMINFO
.set MAGIC,    0x1BADB002
.set CHECKSUM, -(MAGIC + FLAGS)

.section .multiboot
.align 4
.long MAGIC
.long FLAGS
.long CHECKSUM

.section .bss
.align 16
stack_bottom:
.skip 16384
stack_top:

.section .text
.global _start
.type _start, @function
_start:
    mov $stack_top, %esp
    /* push multiboot magic + info for kernel_main */
    push %ebx
    push %eax
    call kernel_main
.hang:
    cli
    hlt
    jmp .hang
.size _start, . - _start
