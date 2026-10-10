/* AlsetOS Kernel — fase 1: boot + VGA + OCB mínimo (sin SO anfitrión) */
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;

#define VGA ((volatile u16*)0xB8000)
#define COLS 80
#define ROWS 25

static u8 cursor_x, cursor_y;

static void clear(void) {
    for (int i = 0; i < COLS * ROWS; i++)
        VGA[i] = (u16)(0x0F << 8) | ' ';
    cursor_x = cursor_y = 0;
}

static void putc(char c) {
    if (c == '\n') {
        cursor_x = 0;
        if (++cursor_y >= ROWS) cursor_y = 0;
        return;
    }
    VGA[cursor_y * COLS + cursor_x] = (u16)(0x0A << 8) | (u8)c;
    if (++cursor_x >= COLS) {
        cursor_x = 0;
        if (++cursor_y >= ROWS) cursor_y = 0;
    }
}

static void print(const char *s) {
    while (*s) putc(*s++);
}

/* Organism Control Block — unidad fundamental (ATS / paradigma Alset) */
struct OCB {
    u32 magic;          /* 'OCB\0' */
    u32 oid_hash;
    u32 phase;          /* 0 boot 1 operacion */
    u32 caps;           /* bitmask capabilities */
    u32 pulse_count;
};

static struct OCB master_ocb;

static u32 hash_str(const char *s) {
    u32 h = 2166136261u;
    while (*s) {
        h ^= (u8)*s++;
        h *= 16777619u;
    }
    return h;
}

void kernel_main(u32 magic, u32 mb_info) {
    (void)mb_info;
    clear();
    print("AlsetOS Kernel Native\n");
    print("=====================\n");
    print("Paradigma: Organismo (no proceso)\n");
    print("Fase 1: boot + VGA + OCB Master\n\n");

    if (magic != 0x2BADB002) {
        print("AVISO: multiboot magic inesperado\n");
    } else {
        print("Multiboot: OK\n");
    }

    master_ocb.magic = 0x4F434200; /* OCB */
    master_ocb.oid_hash = hash_str("Master");
    master_ocb.phase = 1;
    master_ocb.caps = 0xFFFFFFFFu; /* Master: soberano */
    master_ocb.pulse_count = 0;

    print("Master OCB inicializado\n");
    print("OID hash Master listo\n");
    print("Caps: soberanas\n");
    print("\nSin Linux. Sin Windows. Nucleo propio.\n");
    print("Siguiente: scheduler OCB, Pulse, persistencia.\n");
    print("\n[kernel idle — QEMU: cerrar ventana para salir]\n");

    for (;;) {
        __asm__ __volatile__("hlt");
    }
}
