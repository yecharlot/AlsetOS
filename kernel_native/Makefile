# AlsetOS kernel_native
# ARCH=32 (default): ELF32 multiboot → qemu -kernel (funciona)
# ARCH=64: ELF64 long mode → solo via ISO/GRUB (QEMU -kernel no carga ELF64)

ARCH ?= 32
CC = gcc

ifeq ($(ARCH),64)
  CFLAGS = -m64 -ffreestanding -fno-stack-protector -fno-pic -mno-red-zone -O2 -Wall -Wextra
  ASFLAGS = -m64
  LDFLAGS = -m64 -ffreestanding -nostdlib -static -Wl,-T,boot/linker.ld -Wl,--build-id=none
  BOOT_SRC = boot/boot.S
  QEMU = qemu-system-x86_64
else
  CFLAGS = -m32 -ffreestanding -fno-stack-protector -fno-pic -O2 -Wall -Wextra
  ASFLAGS = -m32
  LDFLAGS = -m32 -ffreestanding -nostdlib -static -Wl,-T,boot/linker.ld -Wl,--build-id=none
  BOOT_SRC = boot/boot32.S
  QEMU = qemu-system-x86_64
endif

OBJS = boot/boot.o src/kernel.o

.PHONY: all clean run run64 iso

all: build/alsetos-kernel.bin

build:
	mkdir -p build

boot/boot.o: $(BOOT_SRC) | build
	gcc $(ASFLAGS) -c $(BOOT_SRC) -o $@

src/kernel.o: src/kernel.c | build
	$(CC) $(CFLAGS) -c $< -o $@

build/alsetos-kernel.bin: $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $(OBJS)
	@echo "OK $@ (ARCH=$(ARCH))"

# Por defecto 32-bit: QEMU -kernel solo acepta imagen multiboot 32-bit
run: ARCH=32
run: clean all
	@echo "QEMU -kernel (ELF32 multiboot). Fuente legible + raton estable."
	$(QEMU) -kernel build/alsetos-kernel.bin -m 256M -vga std \
	  -display gtk 2>/dev/null \
	|| $(QEMU) -kernel build/alsetos-kernel.bin -m 256M -vga std \
	  -display sdl 2>/dev/null \
	|| $(QEMU) -kernel build/alsetos-kernel.bin -m 256M -vga std

# 64-bit real: ISO con GRUB
run64: iso
	$(QEMU) -cdrom build/AlsetOS.iso -m 256M -vga std \
	  -display gtk 2>/dev/null \
	|| $(QEMU) -cdrom build/AlsetOS.iso -m 256M -vga std

iso:
	$(MAKE) clean all ARCH=64
	@command -v grub-mkrescue >/dev/null 2>&1 || { \
	  echo "Instala: sudo apt install -y grub-pc-bin grub-common xorriso"; exit 1; }
	mkdir -p build/isodir/boot/grub
	cp build/alsetos-kernel.bin build/isodir/boot/alsetos-kernel.bin
	printf 'set timeout=2\nset default=0\nmenuentry "AlsetOS x86_64" {\n  multiboot /boot/alsetos-kernel.bin\n  boot\n}\n' \
	  > build/isodir/boot/grub/grub.cfg
	grub-mkrescue -o build/AlsetOS.iso build/isodir
	@echo "ISO: build/AlsetOS.iso"

clean:
	rm -f boot/boot.o src/kernel.o
	rm -rf build
