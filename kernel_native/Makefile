CC = gcc
CFLAGS = -m32 -ffreestanding -fno-stack-protector -fno-pic -O2 -Wall -Wextra
LDFLAGS = -m32 -ffreestanding -nostdlib -static -Wl,-T,boot/linker.ld -Wl,--build-id=none
OBJS = boot/boot.o src/kernel.o

.PHONY: all clean run run-gtk iso

all: build/alsetos-kernel.bin

build:
	mkdir -p build

boot/boot.o: boot/boot.S | build
	gcc -m32 -c $< -o $@

src/kernel.o: src/kernel.c | build
	$(CC) $(CFLAGS) -c $< -o $@

build/alsetos-kernel.bin: $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $(OBJS)
	@echo "OK $@"

# Multiboot32: usar i386. x86_64 -kernel falla con "PVH ELF Note".
run: build/alsetos-kernel.bin
	@echo "Controles: flechas, Enter, 1=ORGES 2=Peer 3=Studio tick 4=broadcast"
	@if command -v qemu-system-i386 >/dev/null 2>&1; then \
	  qemu-system-i386 -kernel build/alsetos-kernel.bin -m 64M -display gtk 2>/dev/null \
	  || qemu-system-i386 -kernel build/alsetos-kernel.bin -m 64M -display curses 2>/dev/null \
	  || qemu-system-i386 -kernel build/alsetos-kernel.bin -m 64M -nographic -serial mon:stdio; \
	elif command -v qemu-system-x86_64 >/dev/null 2>&1; then \
	  echo "No hay qemu-system-i386; generando ISO GRUB para x86_64..."; \
	  $(MAKE) iso; \
	  qemu-system-x86_64 -cdrom build/AlsetOS-Kernel.iso -m 64M -display gtk 2>/dev/null \
	  || qemu-system-x86_64 -cdrom build/AlsetOS-Kernel.iso -m 64M; \
	else \
	  echo "Instala QEMU:"; \
	  echo "  sudo apt install -y qemu-system-x86"; \
	  exit 1; \
	fi

run-gtk: build/alsetos-kernel.bin
	qemu-system-i386 -kernel build/alsetos-kernel.bin -m 64M -display gtk

iso: build/alsetos-kernel.bin
	@command -v grub-mkrescue >/dev/null 2>&1 || { echo "sudo apt install -y grub-pc-bin grub-common xorriso"; exit 1; }
	mkdir -p build/isodir/boot/grub
	cp build/alsetos-kernel.bin build/isodir/boot/alsetos-kernel.bin
	printf 'set timeout=0\nset default=0\nmenuentry "AlsetOS Kernel" {\n  multiboot /boot/alsetos-kernel.bin\n  boot\n}\n' > build/isodir/boot/grub/grub.cfg
	grub-mkrescue -o build/AlsetOS-Kernel.iso build/isodir
	@echo "ISO: build/AlsetOS-Kernel.iso"

clean:
	rm -f boot/boot.o src/kernel.o
	rm -rf build
