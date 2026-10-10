CC = gcc
AS = gcc
CFLAGS = -m64 -ffreestanding -fno-stack-protector -fno-pic -mno-red-zone -O2 -Wall -Wextra
ASFLAGS = -m64
LDFLAGS = -m64 -ffreestanding -nostdlib -static -Wl,-T,boot/linker.ld -Wl,--build-id=none
OBJS = boot/boot.o src/kernel.o

.PHONY: all clean run run-curses iso

all: build/alsetos-kernel.bin

build:
	mkdir -p build

boot/boot.o: boot/boot.S | build
	$(AS) $(ASFLAGS) -c $< -o $@

src/kernel.o: src/kernel.c | build
	$(CC) $(CFLAGS) -c $< -o $@

build/alsetos-kernel.bin: $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $(OBJS)
	@echo "OK $@ (x86_64)"

run: build/alsetos-kernel.bin
	@echo "AlsetOS x86_64 — -vga std"
	qemu-system-x86_64 -kernel build/alsetos-kernel.bin -m 256M -vga std \
	  -display gtk 2>/dev/null \
	|| qemu-system-x86_64 -kernel build/alsetos-kernel.bin -m 256M -vga std \
	  -display sdl 2>/dev/null \
	|| qemu-system-x86_64 -kernel build/alsetos-kernel.bin -m 256M -vga std

run-curses: build/alsetos-kernel.bin
	qemu-system-x86_64 -kernel build/alsetos-kernel.bin -m 256M -vga std -display curses

iso: build/alsetos-kernel.bin
	@command -v grub-mkrescue >/dev/null 2>&1 || { echo "need grub-mkrescue"; exit 1; }
	mkdir -p build/isodir/boot/grub
	cp build/alsetos-kernel.bin build/isodir/boot/alsetos-kernel.bin
	printf 'set timeout=3\nset default=0\nmenuentry "AlsetOS x64" {\n  multiboot /boot/alsetos-kernel.bin\n  boot\n}\n' > build/isodir/boot/grub/grub.cfg
	grub-mkrescue -o build/AlsetOS.iso build/isodir

clean:
	rm -f boot/boot.o src/kernel.o
	rm -rf build
