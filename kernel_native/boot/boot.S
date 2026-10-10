/* Multiboot + transición a long mode (x86_64) — AlsetOS */
.set ALIGN,    1<<0
.set MEMINFO,  1<<1
.set VIDEO,    1<<2
.set FLAGS,    ALIGN | MEMINFO | VIDEO
.set MAGIC,    0x1BADB002
.set CHECKSUM, -(MAGIC + FLAGS)

.section .multiboot
.align 4
.long MAGIC
.long FLAGS
.long CHECKSUM
.long 0
.long 800
.long 600
.long 32

.section .bss
.align 4096
.global pml4
pml4:
    .skip 4096
.global pdpt
pdpt:
    .skip 4096
.global pd
pd:
    .skip 4096
.align 16
stack_bottom:
    .skip 65536
stack_top:
mb_magic:
    .long 0
mb_info:
    .long 0

.section .rodata
.align 16
gdt64:
    .quad 0
    .quad 0x00AF9A000000FFFF   /* code 64 */
    .quad 0x00AF92000000FFFF   /* data 64 */
gdt64_end:
gdt64_ptr:
    .word gdt64_end - gdt64 - 1
    .quad gdt64

.section .text
.code32
.global _start
.type _start, @function
_start:
    mov $stack_top, %esp
    mov %eax, mb_magic
    mov %ebx, mb_info

    /* Identity map primeros 4 GiB con páginas de 2 MiB */
    mov $pdpt, %eax
    or $0x03, %eax
    mov %eax, pml4

    mov $pd, %eax
    or $0x03, %eax
    mov %eax, pdpt

    mov $0, %ecx
1:
    mov %ecx, %eax
    shl $21, %eax          /* 2 MiB * index */
    or $0x83, %eax         /* present|rw|PS */
    mov %eax, pd(,%ecx,8)
    inc %ecx
    cmp $512, %ecx
    jb 1b

    /* CR3 = PML4 */
    mov $pml4, %eax
    mov %eax, %cr3

    /* PAE */
    mov %cr4, %eax
    or $0x20, %eax
    mov %eax, %cr4

    /* EFER.LME */
    mov $0xC0000080, %ecx
    rdmsr
    or $0x100, %eax
    wrmsr

    /* Paging + protected */
    mov %cr0, %eax
    or $0x80000001, %eax
    mov %eax, %cr0

    lgdt gdt64_ptr
    ljmp $0x08, $long_mode

.code64
long_mode:
    mov $0x10, %ax
    mov %ax, %ds
    mov %ax, %es
    mov %ax, %ss
    mov %ax, %fs
    mov %ax, %gs
    mov $stack_top, %rsp

    mov mb_magic, %edi
    mov mb_info, %esi
    call kernel_main

.hang:
    cli
    hlt
    jmp .hang
