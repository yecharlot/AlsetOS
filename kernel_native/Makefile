CC = gcc
CFLAGS = -m32 -ffreestanding -fno-stack-protector -fno-pic -O2 -Wall -Wextra
LDFLAGS = -m32 -ffreestanding -nostdlib -static -Wl,-T,boot/linker.ld
OBJS = boot/boot.o src/kernel.o

.PHONY: all clean run iso

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

run: build/alsetos-kernel.bin
	@echo "Controles: flechas, Enter, 1=ORGES 2=Peer 3=Studio tick 4=broadcast"
	qemu-system-i386 -kernel build/alsetos-kernel.bin -display curses 2>/dev/null || \
	qemu-system-i386 -kernel build/alsetos-kernel.bin 2>/dev/null || \
	qemu-system-x86_64 -kernel build/alsetos-kernel.bin

iso: build/alsetos-kernel.bin
	mkdir -p build/isodir/boot/grub
	cp build/alsetos-kernel.bin build/isodir/boot/alsetos-kernel.bin
	printf 'set timeout=0\nset default=0\nmenuentry "AlsetOS Kernel" {\n  multiboot /boot/alsetos-kernel.bin\n}\n' > build/isodir/boot/grub/grub.cfg
	grub-mkrescue -o build/AlsetOS-Kernel.iso build/isodir

clean:
	rm -f boot/boot.o src/kernel.o
	rm -rf build
