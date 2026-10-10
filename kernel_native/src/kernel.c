/*
 * AlsetOS Kernel Native — integración Desktop + organismos (fase integrada)
 * Sin Linux / TinyCore. Unidad = Organismo (OCB). Descentralizado-first.
 */
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;

#define VGA ((volatile u16*)0xB8000)
#define COLS 80
#define ROWS 25
#define MAX_OCB 16
#define MAX_PULSE 32

/* Colores VGA: atributo = (bg<<4)|fg */
#define ATR_NORM  0x0F
#define ATR_DIM   0x08
#define ATR_ACC   0x0A
#define ATR_TITLE 0x0B
#define ATR_WARN  0x0E
#define ATR_INV   0x1F

static u8 cx, cy;

static void vgaput(int x, int y, char c, u8 atr) {
    if (x < 0 || x >= COLS || y < 0 || y >= ROWS) return;
    VGA[y * COLS + x] = (u16)(atr << 8) | (u8)c;
}

static void clear_atr(u8 atr) {
    for (int i = 0; i < COLS * ROWS; i++)
        VGA[i] = (u16)(atr << 8) | ' ';
    cx = cy = 0;
}

static void putc_atr(char c, u8 atr) {
    if (c == '\n') {
        cx = 0;
        if (++cy >= ROWS) cy = ROWS - 1;
        return;
    }
    vgaput(cx, cy, c, atr);
    if (++cx >= COLS) {
        cx = 0;
        if (++cy >= ROWS) cy = ROWS - 1;
    }
}

static void print_atr(const char *s, u8 atr) {
    while (*s) putc_atr(*s++, atr);
}

static void print(const char *s) { print_atr(s, ATR_NORM); }

static void put_hex(u32 v) {
    const char *h = "0123456789ABCDEF";
    print("0x");
    for (int i = 7; i >= 0; i--)
        putc_atr(h[(v >> (i * 4)) & 0xF], ATR_ACC);
}

/* ——— Organism Control Block ——— */
struct OCB {
    u32 magic;       /* 'OCB\0' */
    u32 oid_hash;
    u32 kind;        /* 1 master 2 desktop 3 studio 4 orges 5 peer */
    u32 phase;
    u32 caps;
    u32 pulse_in;
    u32 pulse_out;
    u8  alive;
    char name[16];
};

struct Pulse {
    u32 from;
    u32 to;
    u32 type;
    u32 nonce;
};

static struct OCB ocbs[MAX_OCB];
static int n_ocb;
static struct Pulse pulses[MAX_PULSE];
static int n_pulse;
static u32 node_id;
static u32 nonce_ctr;
static int selected; /* index in menu */

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
    if (n_pulse >= MAX_PULSE) n_pulse = 0; /* ring */
    pulses[n_pulse].from = ocbs[from_i].oid_hash;
    pulses[n_pulse].to = ocbs[to_i].oid_hash;
    pulses[n_pulse].type = type;
    pulses[n_pulse].nonce = ++nonce_ctr;
    ocbs[from_i].pulse_out++;
    ocbs[to_i].pulse_in++;
    n_pulse++;
}

/* ——— PS/2 keyboard (QEMU) ——— */
static inline u8 inb(u16 port) {
    u8 v;
    __asm__ __volatile__("inb %1, %0" : "=a"(v) : "Nd"(port));
    return v;
}

static int kbd_hit(void) {
    return inb(0x64) & 1;
}

/* scancode set 1 → ascii (muy reducido) */
static char scancode_to_ascii(u8 sc) {
    static const char map[128] = {
        0,0,'1','2','3','4','5','6','7','8','9','0','-','=',0,'\t',
        'q','w','e','r','t','y','u','i','o','p','[',']','\n',0,
        'a','s','d','f','g','h','j','k','l',';','\'','`',0,'\\',
        'z','x','c','v','b','n','m',',','.','/',0,'*',0,' ',
    };
    if (sc >= 128) return 0;
    return map[sc];
}

#define KEY_UP    0x48
#define KEY_DOWN  0x50
#define KEY_ENTER 0x1C
#define KEY_ESC   0x01
#define KEY_1     0x02
#define KEY_2     0x03
#define KEY_3     0x04
#define KEY_4     0x05
#define KEY_5     0x06

/* ——— Desktop UI (VGA) ——— */
static void draw_bar(int y, const char *title) {
    for (int x = 0; x < COLS; x++) vgaput(x, y, ' ', ATR_INV);
    int i = 0;
    for (int x = 2; title[i] && x < COLS - 2; x++, i++)
        vgaput(x, y, title[i], ATR_INV);
}

static void draw_desktop(void) {
    clear_atr(0x00);
    draw_bar(0, " AlsetOS Kernel  |  Desktop-Organism  |  decentralized-first  |  sin SO anfitrion ");
    print_atr("\n  NodeID ", ATR_DIM);
    put_hex(node_id);
    print_atr("   OCBs ", ATR_DIM);
    putc_atr('0' + (char)n_ocb, ATR_ACC);
    print_atr("   Pulses ", ATR_DIM);
    putc_atr('0' + (char)(n_pulse > 9 ? 9 : n_pulse), ATR_ACC);
    print_atr("\n\n", ATR_NORM);

    print_atr("  Organismos en este nodo (todo es organismo)\n", ATR_TITLE);
    print_atr("  --------------------------------------------\n", ATR_DIM);

    for (int i = 0; i < n_ocb; i++) {
        u8 atr = (i == selected) ? ATR_INV : ATR_NORM;
        char line[80];
        /* manual format */
        int p = 0;
        line[p++] = ' ';
        line[p++] = (i == selected) ? '>' : ' ';
        line[p++] = ' ';
        const char *nm = ocbs[i].name;
        for (int k = 0; nm[k] && p < 18; k++) line[p++] = nm[k];
        while (p < 20) line[p++] = ' ';
        const char *kname = "?";
        if (ocbs[i].kind == 1) kname = "MASTER ";
        else if (ocbs[i].kind == 2) kname = "DESKTOP";
        else if (ocbs[i].kind == 3) kname = "STUDIO ";
        else if (ocbs[i].kind == 4) kname = "ORGES  ";
        else if (ocbs[i].kind == 5) kname = "PEER   ";
        for (int k = 0; kname[k] && p < 28; k++) line[p++] = kname[k];
        while (p < 30) line[p++] = ' ';
        line[p++] = 'P';
        line[p++] = 'i';
        line[p++] = 'n';
        line[p++] = ':';
        line[p++] = '0' + (ocbs[i].pulse_in > 9 ? 9 : (char)ocbs[i].pulse_in);
        line[p++] = ' ';
        line[p++] = 'o';
        line[p++] = 'u';
        line[p++] = 't';
        line[p++] = ':';
        line[p++] = '0' + (ocbs[i].pulse_out > 9 ? 9 : (char)ocbs[i].pulse_out);
        line[p] = 0;
        for (int x = 0; line[x]; x++) vgaput(x, 6 + i, line[x], atr);
    }

    int base = 6 + n_ocb + 1;
    draw_bar(ROWS - 4, " Teclas ");
    for (int x = 0; x < COLS; x++) vgaput(x, ROWS - 3, ' ', 0x07);
    for (int x = 0; x < COLS; x++) vgaput(x, ROWS - 2, ' ', 0x07);
    const char *help1 = " Up/Down seleccionar | Enter Pulse al seleccionado | 1=ORGES 2=PEER 3=tick Studio";
    const char *help2 = " 4=broadcast Pulse desktop->all | 5=info | ESC=redibujar | QEMU: cerrar para salir";
    for (int i = 0; help1[i] && i < COLS; i++) vgaput(i, ROWS - 3, help1[i], ATR_DIM);
    for (int i = 0; help2[i] && i < COLS; i++) vgaput(i, ROWS - 2, help2[i], ATR_DIM);
    (void)base;
}

static void boot_organisms(void) {
    node_id = fnv("AlsetOS-Node") ^ 0xA15E7001u;
    n_ocb = 0;
    n_pulse = 0;
    nonce_ctr = 0;
    selected = 0;
    /* Caps: bit flags locales — soberanía del nodo, sin autoridad central */
    ocb_add("Master",  1, 0xFFFFFFFFu);
    ocb_add("Desktop", 2, 0x0000FFFFu);
    ocb_add("Studio",  3, 0x0000FF00u);
    ocb_add("PulseBus",4, 0x000000FFu); /* ORGES infraestructura */
}

void kernel_main(u32 magic, u32 mb_info) {
    (void)mb_info;
    boot_organisms();
    draw_desktop();

    if (magic != 0x2BADB002) {
        /* still run — some loaders differ */
    }

    for (;;) {
        if (!kbd_hit()) {
            __asm__ __volatile__("hlt");
            continue;
        }
        u8 sc = inb(0x60);
        if (sc & 0x80) continue; /* break codes */

        if (sc == KEY_UP) {
            if (selected > 0) selected--;
            draw_desktop();
        } else if (sc == KEY_DOWN) {
            if (selected < n_ocb - 1) selected++;
            draw_desktop();
        } else if (sc == KEY_ENTER) {
            /* Pulse Master -> selected (descentralizado: firmado solo por hashes locales) */
            pulse_send(0, (u32)selected, 1);
            draw_desktop();
        } else if (sc == KEY_1) {
            ocb_add("ORGES", 4, 0x00000F00u);
            draw_desktop();
        } else if (sc == KEY_2) {
            ocb_add("Peer", 5, 0x0000000Fu);
            draw_desktop();
        } else if (sc == KEY_3) {
            /* Studio tick: pulse Desktop -> Studio */
            pulse_send(1, 2, 2);
            draw_desktop();
        } else if (sc == KEY_4) {
            for (int i = 0; i < n_ocb; i++)
                pulse_send(1, (u32)i, 3);
            draw_desktop();
        } else if (sc == KEY_5) {
            /* noop redraw with same state */
            draw_desktop();
        } else if (sc == KEY_ESC) {
            draw_desktop();
        }
        (void)scancode_to_ascii;
    }
}
