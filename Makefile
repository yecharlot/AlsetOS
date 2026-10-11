
# —— Desktop moderno (organismo Surface, no bare-metal) ——
.PHONY: desktop desktop-build kernel-desktop

desktop-build:
	@mkdir -p bin
	go build -o bin/alset-desktop-bridge ./cmd/alset-desktop-bridge
	@echo "OK bin/alset-desktop-bridge"

# Gestor moderno: shell HTML + API /v1 · Desktop/Surface/Compositor son organismos
desktop: desktop-build
	@echo ""
	@echo "  Alset Desktop (moderno) — http://127.0.0.1:7420/"
	@echo "  Shell: desktop/shell  ·  Surface organism backend"
	@echo "  NO es el kernel bare-metal (kernel_native)."
	@echo ""
	./bin/alset-desktop-bridge -addr 127.0.0.1:7420 -shell desktop/shell

# Bare-metal (QEMU). Si ves UI vieja: make -C kernel_native clean && make -C kernel_native run
kernel-desktop:
	$(MAKE) -C kernel_native clean
	$(MAKE) -C kernel_native run
