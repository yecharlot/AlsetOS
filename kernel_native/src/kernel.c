/*
 * AlsetOS Kernel Native — Desktop + organismos (teclado PS/2 fiable en QEMU)
 * Sin Linux. Unidad = OCB. Descentralizado-first.
 */
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;

#define VGA ((volatile u16*)0xB8000)
#define COLS 80
#define ROWS 25
#define MAX_OCB 16
#define MAX_PULSE 32

#define ATR_NORM  0x0F
#define ATR_DIM   0x08
#define ATR_ACC   0x0A
#define ATR_TITLE 0x0B
#define ATR_INV   0x1F

static void vgaput(int x, int y, char c, u8 atr) {
    if (x < 0 || x >= COLS || y < 0 || y >= ROWS) return;
    VGA[y * COLS + x] = (u16)(atr << 8) | (u8)c;
}

static void clear_atr(u8 atr) {
    for (int i = 0; i < COLS * ROWS; i++)
        VGA[i] = (u16)(atr << 8) | ' ';
}

static void put_hex_at(int x, int y, u32 v, u8 atr) {
    const char *h = "0123456789ABCDEF";
    vgaput(x, y, '0', atr);
    vgaput(x + 1, y, 'x', atr);
    for (int i = 0; i < 8; i++)
        vgaput(x + 2 + i, y, h[(v >> (28 - i * 4)) & 0xF], atr);
}

struct OCB {
    u32 magic;
    u32 oid_hash;
    u32 kind; /* 1 master 2 desktop 3 studio 4 orges 5 peer */
    u32 phase;
    u32 caps;
    u32 pulse_in;
    u32 pulse_out;
    u8  alive;
    char name[16];
};

struct Pulse {
    u32 from, to, type, nonce;
};

static struct OCB ocbs[MAX_OCB];
static int n_ocb;
static struct Pulse pulses[MAX_PULSE];
static int n_pulse;
static u32 node_id;
static u32 nonce_ctr;
static int selected;
static u32 key_events; /* debug counter on screen */

static u32 fnv(const char *s) {
    u32 h = 2166136261u;
    while (*s) {
        h ^= (u8)*s++;
        h *= 16777619u;
    }
    return h;
}

static void name_copy(char *dst, const char *src) {
    int i = 0;
    for (; i < 15 && src[i]; i++) dst[i] = src[i];
    dst[i] = 0;
}

static int ocb_add(const char *name, u32 kind, u32 caps) {
    if (n_ocb >= MAX_OCB) return -1;
    struct OCB *o = &ocbs[n_ocb];
    o->magic = 0x4F434200;
    o->oid_hash = fnv(name) ^ node_id;
    o->kind = kind;
    o->phase = 1;
    o->caps = caps;
    o->pulse_in = o->pulse_out = 0;
    o->alive = 1;
    name_copy(o->name, name);
    return n_ocb++;
}

static void pulse_send(u32 from_i, u32 to_i, u32 type) {
    if (from_i >= (u32)n_ocb || to_i >= (u32)n_ocb) return;
    if (n_pulse >= MAX_PULSE) n_pulse = 0;
    pulses[n_pulse].from = ocbs[from_i].oid_hash;
    pulses[n_pulse].to = ocbs[to_i].oid_hash;
    pulses[n_pulse].type = type;
    pulses[n_pulse].nonce = ++nonce_ctr;
    ocbs[from_i].pulse_out++;
    ocbs[to_i].pulse_in++;
    n_pulse++;
}

/* ——— I/O ——— */
static inline u8 inb(u16 port) {
    u8 v;
    __asm__ __volatile__("inb %1, %0" : "=a"(v) : "Nd"(port));
    return v;
}

static inline void outb(u16 port, u8 val) {
    __asm__ __volatile__("outb %0, %1" : : "a"(val), "Nd"(port));
}

/* Inicializar controlador 8042 / PS/2 (QEMU) */
static void kbd_init(void) {
    /* vaciar buffer de salida */
    for (int i = 0; i < 256; i++) {
        if (!(inb(0x64) & 1)) break;
        (void)inb(0x60);
    }
    /* habilitar teclado: command byte */
    outb(0x64, 0xAE); /* enable keyboard interface */
}

/* Espera activa breve (sin hlt: hlt sin IRQ deja el teclado muerto) */
static void pause_poll(void) {
    for (volatile int i = 0; i < 5000; i++)
        __asm__ __volatile__("pause");
}

static int kbd_read_scancode(u8 *out) {
    if (!(inb(0x64) & 1))
        return 0;
    *out = inb(0x60);
    return 1;
}

#define SC_EXT    0xE0
#define SC_UP     0x48
#define SC_DOWN   0x50
#define SC_ENTER  0x1C
#define SC_ESC    0x01
#define SC_1      0x02
#define SC_2      0x03
#define SC_3      0x04
#define SC_4      0x05
#define SC_5      0x06
#define SC_W      0x11
#define SC_S      0x1F
#define SC_SPACE  0x39

static void draw_bar(int y, const char *title) {
    for (int x = 0; x < COLS; x++) vgaput(x, y, ' ', ATR_INV);
    for (int i = 0, x = 2; title[i] && x < COLS - 2; x++, i++)
        vgaput(x, y, title[i], ATR_INV);
}

static void draw_desktop(void) {
    clear_atr(0x00);
    draw_bar(0, " AlsetOS Kernel | Desktop-Organism | decentralized-first | sin SO anfitrion ");

    const char *l1 = " NodeID ";
    for (int i = 0; l1[i]; i++) vgaput(2 + i, 2, l1[i], ATR_DIM);
    put_hex_at(10, 2, node_id, ATR_ACC);

    const char *l2 = " OCBs ";
    for (int i = 0; l2[i]; i++) vgaput(22 + i, 2, l2[i], ATR_DIM);
    vgaput(28, 2, (char)('0' + (n_ocb > 9 ? 9 : n_ocb)), ATR_ACC);

    const char *l3 = " Pulses ";
    for (int i = 0; l3[i]; i++) vgaput(32 + i, 2, l3[i], ATR_DIM);
    vgaput(40, 2, (char)('0' + (n_pulse > 9 ? 9 : n_pulse)), ATR_ACC);

    const char *l4 = " Keys ";
    for (int i = 0; l4[i]; i++) vgaput(44 + i, 2, l4[i], ATR_DIM);
    put_hex_at(50, 2, key_events, ATR_ACC);

    const char *title = " Organismos en este nodo (todo es organismo)";
    for (int i = 0; title[i] && i < COLS; i++) vgaput(2 + i, 4, title[i], ATR_TITLE);
    for (int x = 2; x < 50; x++) vgaput(x, 5, '-', ATR_DIM);

    for (int i = 0; i < n_ocb; i++) {
        u8 atr = (i == selected) ? ATR_INV : ATR_NORM;
        int y = 6 + i;
        vgaput(2, y, (i == selected) ? '>' : ' ', atr);
        const char *nm = ocbs[i].name;
        int x = 4;
        for (int k = 0; nm[k] && x < 18; k++, x++) vgaput(x, y, nm[k], atr);
        while (x < 20) vgaput(x++, y, ' ', atr);

        const char *kname = "?????";
        if (ocbs[i].kind == 1) kname = "MASTER ";
        else if (ocbs[i].kind == 2) kname = "DESKTOP";
        else if (ocbs[i].kind == 3) kname = "STUDIO ";
        else if (ocbs[i].kind == 4) kname = "ORGES  ";
        else if (ocbs[i].kind == 5) kname = "PEER   ";
        for (int k = 0; kname[k]; k++) vgaput(20 + k, y, kname[k], atr);

        vgaput(30, y, 'P', atr);
        vgaput(31, y, 'i', atr);
        vgaput(32, y, 'n', atr);
        vgaput(33, y, ':', atr);
        vgaput(34, y, (char)('0' + (ocbs[i].pulse_in > 9 ? 9 : (int)ocbs[i].pulse_in)), atr);
        vgaput(36, y, 'o', atr);
        vgaput(37, y, 'u', atr);
        vgaput(38, y, 't', atr);
        vgaput(39, y, ':', atr);
        vgaput(40, y, (char)('0' + (ocbs[i].pulse_out > 9 ? 9 : (int)ocbs[i].pulse_out)), atr);
    }

    draw_bar(ROWS - 4, " Teclas ");
    const char *h1 = " W/S o flechas: seleccionar | Enter/Espacio: Pulse | 1=ORGES 2=PEER 3=Studio 4=broadcast";
    const char *h2 = " 5=info ESC=redibujar | clic en ventana QEMU para foco del teclado | cerrar QEMU para salir";
    for (int i = 0; h1[i] && i < COLS; i++) vgaput(i, ROWS - 3, h1[i], ATR_DIM);
    for (int i = 0; h2[i] && i < COLS; i++) vgaput(i, ROWS - 2, h2[i], ATR_DIM);
}

static void boot_organisms(void) {
    node_id = fnv("AlsetOS-Node") ^ 0xA15E7001u;
    n_ocb = 0;
    n_pulse = 0;
    nonce_ctr = 0;
    selected = 0;
    key_events = 0;
    ocb_add("Master",  1, 0xFFFFFFFFu);
    ocb_add("Desktop", 2, 0x0000FFFFu);
    ocb_add("Studio",  3, 0x0000FF00u);
    ocb_add("PulseBus",4, 0x000000FFu);
}

static void handle_key(u8 sc, int extended) {
    key_events++;
    /* break codes */
    if (sc & 0x80)
        return;

    if (sc == SC_UP || (extended && sc == SC_UP) || sc == SC_W) {
        if (selected > 0) selected--;
        draw_desktop();
        return;
    }
    if (sc == SC_DOWN || (extended && sc == SC_DOWN) || sc == SC_S) {
        if (selected < n_ocb - 1) selected++;
        draw_desktop();
        return;
    }
    if (sc == SC_ENTER || sc == SC_SPACE) {
        pulse_send(0, (u32)selected, 1);
        draw_desktop();
        return;
    }
    if (sc == SC_1) {
        ocb_add("ORGES", 4, 0x00000F00u);
        draw_desktop();
        return;
    }
    if (sc == SC_2) {
        ocb_add("Peer", 5, 0x0000000Fu);
        draw_desktop();
        return;
    }
    if (sc == SC_3) {
        pulse_send(1, 2, 2);
        draw_desktop();
        return;
    }
    if (sc == SC_4) {
        for (int i = 0; i < n_ocb; i++)
            pulse_send(1, (u32)i, 3);
        draw_desktop();
        return;
    }
    if (sc == SC_5 || sc == SC_ESC) {
        draw_desktop();
        return;
    }
    /* tecla desconocida: al menos actualiza contador Keys */
    draw_desktop();
}

void kernel_main(u32 magic, u32 mb_info) {
    (void)magic;
    (void)mb_info;

    kbd_init();
    boot_organisms();
    draw_desktop();

    int ext = 0;
    for (;;) {
        u8 sc;
        if (!kbd_read_scancode(&sc)) {
            pause_poll();
            continue;
        }
        if (sc == SC_EXT) {
            ext = 1;
            continue;
        }
        handle_key(sc, ext);
        ext = 0;
    }
}
