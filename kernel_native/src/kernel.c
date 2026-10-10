
/*
 * AlsetOS Kernel x86_64 nativo
 * Maestro → Framebuffer → Compositor → Desktop → Ventana/Texto/Lista
 */
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long u64;

#define MAX_OCB 48
#define MAX_NAME 16
#define MAX_TITLE 28
#define MAX_TEXT 96
#define MAX_LIST 6
#define MAX_LIST_ITEM 20

#define CAP_PULSO_ENV   0x01u
#define CAP_PULSO_REC   0x02u
#define CAP_ORGES_CREAR 0x04u
#define CAP_ZYRION      0x40u
#define CAP_DISP_LEER   0x80u
#define CAP_DISP_ESCR   0x100u
#define CAP_DISP_CTRL   0x200u
#define CAP_DESKTOP     0x400u
#define CAP_COMPOSE     0x800u
#define CAP_FB          0x1000u
#define CAP_ALL         0xFFFFFFFFu

#define K_MASTER  1
#define K_SHELL   2
#define K_ORGES   3
#define K_BUS     5
#define K_DISP    6
#define K_DESKTOP 7
#define K_FB      8
#define K_COMPOSE 9
#define K_WINDOW  10
#define K_TEXT    11
#define K_LIST    12

#define COL_BG     0xFF0B1220u
#define COL_PANEL  0xFF151E32u
#define COL_TASK   0xFF1A2744u
#define COL_ACCENT 0xFF3DDC97u
#define COL_TITLE  0xFF5B8DEEu
#define COL_TEXT   0xFFE8EEF8u
#define COL_DIM    0xFF8A9BB5u
#define COL_WIN_BG 0xFF1C2538u
#define COL_WIN_TOP 0xFF243044u
#define COL_DANGER 0xFFE05C5Cu
#define COL_CURSOR 0xFFFFFFFFu
#define COL_ICON   0xFF4FC3F7u
#define COL_MENU   0xFF1E2A40u
#define COL_SEL    0xFF2E4A6Eu

static inline u8 inb(u16 p){u8 v;__asm__ __volatile__("inb %1,%0":"=a"(v):"Nd"(p));return v;}
static inline void outb(u16 p,u8 v){__asm__ __volatile__("outb %0,%1"::"a"(v),"Nd"(p));}
static inline void outw(u16 p,u16 v){__asm__ __volatile__("outw %0,%1"::"a"(v),"Nd"(p));}
static inline u16 inw(u16 p){u16 v;__asm__ __volatile__("inw %1,%0":"=a"(v):"Nd"(p));return v;}
static void pausa(void){for(volatile int i=0;i<400;i++)__asm__ __volatile__("pause");}

struct OCB {
    u32 magic, oid, kind, caps, pin, pout;
    u8 alive, visible, focused;
    char nombre[MAX_NAME];
    char titulo[MAX_TITLE];
    int x, y, w, h;
    u32 color, disp_clase, disp_dato;
    char text[MAX_TEXT];
    int text_len;
    char items[MAX_LIST][MAX_LIST_ITEM];
    int n_items, sel_item;
};

static struct OCB ocbs[MAX_OCB];
static int n_ocb, actor_i, n_pulso, focus_win;
static u32 nodo_id, nonce;
static u32 *fb;
static u32 fb_w, fb_h, fb_pitch;
static int fb_ok;
static int mouse_x = 400, mouse_y = 300;
static u8 mouse_btn, mouse_cycle, mouse_pkt[3];
static int drag_win = -1, drag_ox, drag_oy, menu_open, orges_seq;
static int mouse_dirty;
static volatile u16 *const VGA = (volatile u16 *)0xB8000UL;

static const u8 FONT8[96][8] = {
  {0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00},   {0x18,0x3C,0x3C,0x18,0x18,0x00,0x18,0x00},   {0x36,0x36,0x00,0x00,0x00,0x00,0x00,0x00},   {0x36,0x36,0x7F,0x36,0x7F,0x36,0x36,0x00},
  {0x0C,0x3E,0x03,0x1E,0x30,0x1F,0x0C,0x00},   {0x00,0x63,0x33,0x18,0x0C,0x66,0x63,0x00},   {0x1C,0x36,0x1C,0x6E,0x3B,0x33,0x6E,0x00},   {0x06,0x06,0x03,0x00,0x00,0x00,0x00,0x00},
  {0x18,0x0C,0x06,0x06,0x06,0x0C,0x18,0x00},   {0x06,0x0C,0x18,0x18,0x18,0x0C,0x06,0x00},   {0x00,0x66,0x3C,0xFF,0x3C,0x66,0x00,0x00},   {0x00,0x0C,0x0C,0x3F,0x0C,0x0C,0x00,0x00},
  {0x00,0x00,0x00,0x00,0x00,0x0C,0x0C,0x06},   {0x00,0x00,0x00,0x3F,0x00,0x00,0x00,0x00},   {0x00,0x00,0x00,0x00,0x00,0x0C,0x0C,0x00},   {0x60,0x30,0x18,0x0C,0x06,0x03,0x01,0x00},
  {0x3E,0x63,0x73,0x7B,0x6F,0x67,0x3E,0x00},   {0x0C,0x0E,0x0C,0x0C,0x0C,0x0C,0x3F,0x00},   {0x1E,0x33,0x30,0x1C,0x06,0x33,0x3F,0x00},   {0x1E,0x33,0x30,0x1C,0x30,0x33,0x1E,0x00},
  {0x38,0x3C,0x36,0x33,0x7F,0x30,0x78,0x00},   {0x3F,0x03,0x1F,0x30,0x30,0x33,0x1E,0x00},   {0x1C,0x06,0x03,0x1F,0x33,0x33,0x1E,0x00},   {0x3F,0x33,0x30,0x18,0x0C,0x0C,0x0C,0x00},
  {0x1E,0x33,0x33,0x1E,0x33,0x33,0x1E,0x00},   {0x1E,0x33,0x33,0x3E,0x30,0x18,0x0E,0x00},   {0x00,0x0C,0x0C,0x00,0x00,0x0C,0x0C,0x00},   {0x00,0x0C,0x0C,0x00,0x00,0x0C,0x0C,0x06},
  {0x18,0x0C,0x06,0x03,0x06,0x0C,0x18,0x00},   {0x00,0x00,0x3F,0x00,0x00,0x3F,0x00,0x00},   {0x06,0x0C,0x18,0x30,0x18,0x0C,0x06,0x00},   {0x1E,0x33,0x30,0x18,0x0C,0x00,0x0C,0x00},
  {0x3E,0x63,0x7B,0x7B,0x7B,0x03,0x1E,0x00},   {0x0C,0x1E,0x33,0x33,0x3F,0x33,0x33,0x00},   {0x3F,0x66,0x66,0x3E,0x66,0x66,0x3F,0x00},   {0x3C,0x66,0x03,0x03,0x03,0x66,0x3C,0x00},
  {0x1F,0x36,0x66,0x66,0x66,0x36,0x1F,0x00},   {0x7F,0x46,0x16,0x1E,0x16,0x46,0x7F,0x00},   {0x7F,0x46,0x16,0x1E,0x16,0x06,0x0F,0x00},   {0x3C,0x66,0x03,0x03,0x73,0x66,0x7C,0x00},
  {0x33,0x33,0x33,0x3F,0x33,0x33,0x33,0x00},   {0x1E,0x0C,0x0C,0x0C,0x0C,0x0C,0x1E,0x00},   {0x78,0x30,0x30,0x30,0x33,0x33,0x1E,0x00},   {0x67,0x66,0x36,0x1E,0x36,0x66,0x67,0x00},
  {0x0F,0x06,0x06,0x06,0x46,0x66,0x7F,0x00},   {0x63,0x77,0x7F,0x6B,0x63,0x63,0x63,0x00},   {0x63,0x67,0x6F,0x7B,0x73,0x63,0x63,0x00},   {0x1C,0x36,0x63,0x63,0x63,0x36,0x1C,0x00},
  {0x3F,0x66,0x66,0x3E,0x06,0x06,0x0F,0x00},   {0x1E,0x33,0x33,0x33,0x3B,0x1E,0x38,0x00},   {0x3F,0x66,0x66,0x3E,0x36,0x66,0x67,0x00},   {0x1E,0x33,0x07,0x0E,0x38,0x33,0x1E,0x00},
  {0x3F,0x2D,0x0C,0x0C,0x0C,0x0C,0x1E,0x00},   {0x33,0x33,0x33,0x33,0x33,0x33,0x3F,0x00},   {0x33,0x33,0x33,0x33,0x33,0x1E,0x0C,0x00},   {0x63,0x63,0x63,0x6B,0x7F,0x77,0x63,0x00},
  {0x63,0x63,0x36,0x1C,0x1C,0x36,0x63,0x00},   {0x33,0x33,0x33,0x1E,0x0C,0x0C,0x1E,0x00},   {0x7F,0x63,0x31,0x18,0x4C,0x66,0x7F,0x00},   {0x1E,0x06,0x06,0x06,0x06,0x06,0x1E,0x00},
  {0x03,0x06,0x0C,0x18,0x30,0x60,0x40,0x00},   {0x1E,0x18,0x18,0x18,0x18,0x18,0x1E,0x00},   {0x08,0x1C,0x36,0x63,0x00,0x00,0x00,0x00},   {0x00,0x00,0x00,0x00,0x00,0x00,0x00,0xFF},
  {0x0C,0x0C,0x18,0x00,0x00,0x00,0x00,0x00},   {0x00,0x00,0x1E,0x30,0x3E,0x33,0x6E,0x00},   {0x07,0x06,0x06,0x3E,0x66,0x66,0x3B,0x00},   {0x00,0x00,0x1E,0x33,0x03,0x33,0x1E,0x00},
  {0x38,0x30,0x30,0x3E,0x33,0x33,0x6E,0x00},   {0x00,0x00,0x1E,0x33,0x3F,0x03,0x1E,0x00},   {0x1C,0x36,0x06,0x0F,0x06,0x06,0x0F,0x00},   {0x00,0x00,0x6E,0x33,0x33,0x3E,0x30,0x1F},
  {0x07,0x06,0x36,0x6E,0x66,0x66,0x67,0x00},   {0x0C,0x00,0x0E,0x0C,0x0C,0x0C,0x1E,0x00},   {0x30,0x00,0x30,0x30,0x30,0x33,0x33,0x1E},   {0x07,0x06,0x66,0x36,0x1E,0x36,0x67,0x00},
  {0x0E,0x0C,0x0C,0x0C,0x0C,0x0C,0x1E,0x00},   {0x00,0x00,0x33,0x7F,0x7F,0x6B,0x63,0x00},   {0x00,0x00,0x1F,0x33,0x33,0x33,0x33,0x00},   {0x00,0x00,0x1E,0x33,0x33,0x33,0x1E,0x00},
  {0x00,0x00,0x3B,0x66,0x66,0x3E,0x06,0x0F},   {0x00,0x00,0x6E,0x33,0x33,0x3E,0x30,0x78},   {0x00,0x00,0x3B,0x6E,0x66,0x06,0x0F,0x00},   {0x00,0x00,0x3E,0x03,0x1E,0x30,0x1F,0x00},
  {0x08,0x0C,0x3E,0x0C,0x0C,0x2C,0x18,0x00},   {0x00,0x00,0x33,0x33,0x33,0x33,0x6E,0x00},   {0x00,0x00,0x33,0x33,0x33,0x1E,0x0C,0x00},   {0x00,0x00,0x63,0x6B,0x7F,0x7F,0x36,0x00},
  {0x00,0x00,0x63,0x36,0x1C,0x36,0x63,0x00},   {0x00,0x00,0x33,0x33,0x33,0x3E,0x30,0x1F},   {0x00,0x00,0x3F,0x19,0x0C,0x26,0x3F,0x00},   {0x38,0x0C,0x0C,0x07,0x0C,0x0C,0x38,0x00},
  {0x18,0x18,0x18,0x00,0x18,0x18,0x18,0x00},   {0x07,0x0C,0x0C,0x38,0x0C,0x0C,0x07,0x00},   {0x6E,0x3B,0x00,0x00,0x00,0x00,0x00,0x00},   {0x08,0x1C,0x36,0x63,0x63,0x7F,0x00,0x00},

};


static u32 fnv(const char *s) {
    u32 h = 2166136261u;
    while (*s) { h ^= (u8)*s++; h *= 16777619u; }
    return h;
}
static void cpy(char *d, const char *s, int n) {
    int i = 0; for (; i < n - 1 && s[i]; i++) d[i] = s[i]; d[i] = 0;
}
static int igual(const char *a, const char *b) {
    while (*a && *b && *a == *b) { a++; b++; }
    return !*a && !*b;
}
static int buscar(const char *n) {
    for (int i = 0; i < n_ocb; i++)
        if (ocbs[i].alive && igual(ocbs[i].nombre, n)) return i;
    return -1;
}
static int tiene_cap(int i, u32 c) {
    if (i < 0 || !ocbs[i].alive) return 0;
    if (ocbs[i].kind == K_MASTER) return 1;
    return (ocbs[i].caps & c) != 0;
}
static int ocb_add(const char *n, u32 kind, u32 caps) {
    if (n_ocb >= MAX_OCB) return -1;
    if (buscar(n) >= 0) return -2;
    struct OCB *o = &ocbs[n_ocb];
    o->magic = 0x4F434200;
    o->oid = fnv(n) ^ nodo_id ^ (u32)n_ocb;
    o->kind = kind; o->caps = caps;
    o->pin = o->pout = 0; o->alive = 1;
    o->x = 80; o->y = 60; o->w = 280; o->h = 180;
    o->visible = 1; o->focused = 0;
    o->color = COL_WIN_BG;
    o->disp_clase = o->disp_dato = 0;
    o->text[0] = 0; o->text_len = 0;
    o->n_items = 0; o->sel_item = 0; o->titulo[0] = 0;
    cpy(o->nombre, n, MAX_NAME);
    return n_ocb++;
}
static int pulso(int from, int to) {
    if (from < 0 || to < 0) return -1;
    if (!tiene_cap(from, CAP_PULSO_ENV)) return -2;
    if (!tiene_cap(to, CAP_PULSO_REC)) return -3;
    ocbs[from].pout++; ocbs[to].pin++;
    nonce++; n_pulso++;
    return 0;
}

#define VBE_IDX 0x01CE
#define VBE_DAT 0x01CF
static void vbe_write(u16 i, u16 v) { outw(VBE_IDX, i); outw(VBE_DAT, v); }
static u16 vbe_read(u16 i) { outw(VBE_IDX, i); return inw(VBE_DAT); }

static int fb_probe_addr(u64 addr, u32 pixels) {
    volatile u32 *p = (volatile u32 *)addr;
    p[0] = 0xA11E70ADu;
    p[1] = 0xB01DFACEu;
    if (p[0] != 0xA11E70ADu || p[1] != 0xB01DFACEu) return 0;
    for (u32 i = 0; i < 32 && i < pixels; i++) p[i] = 0xFF3DDC97u;
    return 1;
}

static int fb_try_bases(u32 w, u32 h) {
    static const u64 bases[] = {
        0xE0000000ULL, 0xFD000000ULL, 0xF0000000ULL,
        0xD0000000ULL, 0x80000000ULL, 0
    };
    fb_w = w; fb_h = h; fb_pitch = w * 4;
    u32 pixels = w * h;
    for (int i = 0; bases[i]; i++) {
        if (fb_probe_addr(bases[i], pixels)) {
            fb = (u32 *)bases[i];
            fb_ok = 1;
            return 1;
        }
    }
    return 0;
}

static int fb_init_bochs(int w, int h) {
    vbe_write(0, 0xB0C0);
    if (vbe_read(0) < 0xB0C0) return 0;
    vbe_write(4, 0);
    vbe_write(1, (u16)w);
    vbe_write(2, (u16)h);
    vbe_write(3, 32);
    vbe_write(6, (u16)w);
    vbe_write(7, (u16)h);
    vbe_write(8, 0);
    vbe_write(9, 0);
    vbe_write(4, 0x41);
    return fb_try_bases((u32)w, (u32)h);
}

static int fb_init_multiboot(u32 magic, u32 info) {
    if (magic != 0x2BADB002u || !info) return 0;
    u32 *mi = (u32 *)(u64)info;
    if (!(mi[0] & (1u << 12))) return 0;
    u32 *f = (u32 *)((u64)info + 88);
    u32 addr = f[0];
    fb_pitch = f[2]; fb_w = f[3]; fb_h = f[4];
    u8 bpp = *((u8 *)((u64)info + 108));
    if (!addr || fb_w < 320 || bpp != 32) return 0;
    if (!fb_probe_addr((u64)addr, fb_w * fb_h)) return 0;
    fb = (u32 *)(u64)addr;
    fb_ok = 1;
    return 1;
}

static int fb_init(u32 magic, u32 info) {
    if (fb_init_multiboot(magic, info)) return 1;
    if (fb_init_bochs(800, 600)) return 1;
    if (fb_init_bochs(640, 480)) return 1;
    return 0;
}

static void put_px(int x, int y, u32 c) {
    if (!fb_ok || x < 0 || y < 0 || (u32)x >= fb_w || (u32)y >= fb_h) return;
    ((u32 *)((u8 *)fb + (u32)y * fb_pitch))[x] = c;
}
static void fill_rect(int x, int y, int w, int h, u32 c) {
    if (!fb_ok || w <= 0 || h <= 0) return;
    if (x < 0) { w += x; x = 0; }
    if (y < 0) { h += y; y = 0; }
    if ((u32)(x + w) > fb_w) w = (int)fb_w - x;
    if ((u32)(y + h) > fb_h) h = (int)fb_h - y;
    for (int yy = 0; yy < h; yy++) {
        u32 *row = (u32 *)((u8 *)fb + (u32)(y + yy) * fb_pitch);
        for (int xx = 0; xx < w; xx++) row[x + xx] = c;
    }
}
static void draw_rect(int x, int y, int w, int h, u32 c) {
    fill_rect(x, y, w, 1, c);
    fill_rect(x, y + h - 1, w, 1, c);
    fill_rect(x, y, 1, h, c);
    fill_rect(x + w - 1, y, 1, h, c);
}
static void draw_char(int x, int y, char ch, u32 fg) {
    int idx = (int)(u8)ch - 32;
    if (idx < 0 || idx >= 96) idx = 0;
    const u8 *g = FONT8[idx];
    for (int row = 0; row < 8; row++) {
        u8 bits = g[row];
        for (int col = 0; col < 8; col++)
            if (bits & (1 << col))
                put_px(x + col, y + row, fg);
    }
}
static void draw_text(int x, int y, const char *s, u32 fg) {
    while (*s && (u32)x + 8 <= fb_w) {
        draw_char(x, y, *s++, fg);
        x += 8;
    }
}

static int is_win(int i) {
    u32 k = ocbs[i].kind;
    return k == K_WINDOW || k == K_SHELL || k == K_TEXT || k == K_LIST || k == K_ORGES;
}
static void focus_set(int i) {
    for (int j = 0; j < n_ocb; j++) ocbs[j].focused = 0;
    if (i >= 0 && i < n_ocb && ocbs[i].alive) {
        ocbs[i].focused = 1;
        focus_win = i;
        actor_i = i;
    }
}
static int hit_window(int mx, int my) {
    if (focus_win >= 0 && focus_win < n_ocb && is_win(focus_win)) {
        struct OCB *o = &ocbs[focus_win];
        if (o->alive && o->visible &&
            mx >= o->x && mx < o->x + o->w && my >= o->y && my < o->y + o->h)
            return focus_win;
    }
    for (int i = n_ocb - 1; i >= 0; i--) {
        if (!ocbs[i].alive || !ocbs[i].visible || !is_win(i)) continue;
        struct OCB *o = &ocbs[i];
        if (mx >= o->x && mx < o->x + o->w && my >= o->y && my < o->y + o->h)
            return i;
    }
    return -1;
}

static int crear_orges(u32 kind, const char *base) {
    int di = buscar("Desktop");
    if (di < 0 || !tiene_cap(di, CAP_ORGES_CREAR)) return -1;
    char name[MAX_NAME];
    int n = 0;
    while (base[n] && n < 10) { name[n] = base[n]; n++; }
    name[n++] = (char)('0' + (orges_seq % 9) + 1);
    name[n] = 0;
    orges_seq++;
    int id = ocb_add(name, kind, CAP_PULSO_ENV | CAP_PULSO_REC);
    if (id < 0) return id;
    ocbs[id].x = 100 + (orges_seq * 30) % 200;
    ocbs[id].y = 70 + (orges_seq * 24) % 120;
    ocbs[id].w = (kind == K_LIST) ? 260 : 320;
    ocbs[id].h = (kind == K_TEXT) ? 200 : 180;
    cpy(ocbs[id].titulo, name, MAX_TITLE);
    if (kind == K_TEXT) {
        cpy(ocbs[id].text, "Escribe aqui", MAX_TEXT);
        ocbs[id].text_len = 12;
    }
    if (kind == K_LIST) {
        cpy(ocbs[id].items[0], "item-alpha", MAX_LIST_ITEM);
        cpy(ocbs[id].items[1], "item-beta", MAX_LIST_ITEM);
        cpy(ocbs[id].items[2], "item-gamma", MAX_LIST_ITEM);
        ocbs[id].n_items = 3;
    }
    if (kind == K_WINDOW || kind == K_ORGES)
        cpy(ocbs[id].text, "ORGES nativa x64", MAX_TEXT);
    focus_set(id);
    pulso(di, id);
    return id;
}

static void compose_window(int i) {
    struct OCB *o = &ocbs[i];
    if (!o->alive || !o->visible || !is_win(i)) return;
    int x = o->x, y = o->y, w = o->w, h = o->h;
    if (w < 120) w = 120;
    if (h < 90) h = 90;
    u32 border = o->focused ? COL_ACCENT : COL_TITLE;
    fill_rect(x, y, w, h, o->color);
    fill_rect(x, y, w, 28, COL_WIN_TOP);
    draw_rect(x, y, w, h, border);
    draw_text(x + 10, y + 10, o->titulo[0] ? o->titulo : o->nombre, COL_TEXT);
    fill_rect(x + w - 24, y + 6, 16, 16, COL_DANGER);
    draw_text(x + w - 20, y + 10, "x", COL_TEXT);

    if (o->kind == K_TEXT) {
        fill_rect(x + 10, y + 40, w - 20, h - 60, COL_PANEL);
        draw_text(x + 16, y + 48, o->text, COL_TEXT);
        if (o->focused)
            draw_text(x + 16 + o->text_len * 8, y + 48, "_", COL_ACCENT);
        draw_text(x + 16, y + h - 24, "texto editable", COL_DIM);
    } else if (o->kind == K_LIST) {
        draw_text(x + 12, y + 40, "Lista ORGES", COL_DIM);
        for (int k = 0; k < o->n_items; k++) {
            if (k == o->sel_item)
                fill_rect(x + 10, y + 56 + k * 18, w - 20, 16, COL_SEL);
            draw_text(x + 16, y + 58 + k * 18, o->items[k],
                      (k == o->sel_item && o->focused) ? COL_ACCENT : COL_TEXT);
        }
    } else if (o->kind == K_SHELL) {
        draw_text(x + 12, y + 48, "Consola nativa", COL_DIM);
        draw_text(x + 12, y + 68, "AlsetOS x86_64", COL_TEXT);
        draw_text(x + 12, y + 96, "M = menu Desktop", COL_DIM);
    } else {
        draw_text(x + 12, y + 48, "ORGES nativa", COL_DIM);
        draw_text(x + 12, y + 68, o->text[0] ? o->text : o->nombre, COL_ACCENT);
        draw_text(x + 12, y + 96, "Pulso y caps activos", COL_DIM);
    }
}

static void compositor_tick(void) {
    if (!fb_ok) return;
    int ci = buscar("Compositor");
    if (ci >= 0) ocbs[ci].disp_dato++;

    fill_rect(0, 0, (int)fb_w, (int)fb_h, COL_BG);
    fill_rect(0, 0, (int)fb_w, 40, COL_PANEL);
    draw_text(16, 14, "AlsetOS Desktop x64", COL_ACCENT);
    draw_text(200, 14, "Maestro > FB > Compositor > Desktop", COL_DIM);

    int ix = 20, iy = 56;
    for (int i = 0; i < n_ocb; i++) {
        if (!ocbs[i].alive || ocbs[i].kind != K_DISP) continue;
        fill_rect(ix, iy, 80, 56, COL_PANEL);
        draw_rect(ix, iy, 80, 56, COL_ICON);
        draw_text(ix + 8, iy + 14, ocbs[i].nombre, COL_TEXT);
        draw_text(ix + 8, iy + 32, "DISP", COL_DIM);
        iy += 68;
        if (iy > (int)fb_h - 100) { iy = 56; ix += 96; }
    }

    for (int i = 0; i < n_ocb; i++)
        if (i != focus_win) compose_window(i);
    if (focus_win >= 0) compose_window(focus_win);

    if (menu_open) {
        fill_rect(12, 48, 220, 150, COL_MENU);
        draw_rect(12, 48, 220, 150, COL_ACCENT);
        draw_text(24, 60, "Menu Desktop", COL_ACCENT);
        draw_text(24, 88, "1  Nueva ventana", COL_TEXT);
        draw_text(24, 108, "2  Texto editable", COL_TEXT);
        draw_text(24, 128, "3  Lista", COL_TEXT);
        draw_text(24, 148, "Enter  Pulso", COL_DIM);
        draw_text(24, 168, "M  Cerrar menu", COL_DIM);
    }

    int th = 44, ty = (int)fb_h - th;
    fill_rect(0, ty, (int)fb_w, th, COL_TASK);
    fill_rect(0, ty, (int)fb_w, 3, COL_ACCENT);
    draw_text(16, ty + 16, "Desktop", COL_ACCENT);
    int sx = 100;
    for (int i = 0; i < n_ocb; i++) {
        if (!ocbs[i].alive || !is_win(i)) continue;
        draw_text(sx, ty + 16, ocbs[i].nombre,
                  (i == focus_win) ? COL_ACCENT : COL_DIM);
        sx += 90;
        if (sx > (int)fb_w - 180) break;
    }
    draw_text((int)fb_w - 220, ty + 16, "M menu | 1/2/3 ORGES", COL_DIM);

    /* cursor estable */
    for (int i = 0; i < 14; i++) {
        put_px(mouse_x, mouse_y + i, COL_CURSOR);
        put_px(mouse_x + 1, mouse_y + i, COL_CURSOR);
    }
    for (int i = 0; i < 10; i++)
        put_px(mouse_x + i, mouse_y + i, COL_ACCENT);
}

static void mouse_init(void) {
    /* reset auxiliar */
    outb(0x64, 0xA8);
    outb(0x64, 0x20);
    u8 s = inb(0x60);
    s |= 0x02;
    s &= (u8)~0x20;
    outb(0x64, 0x60);
    outb(0x60, s);
    outb(0x64, 0xD4);
    outb(0x60, 0xF6); /* defaults */
    /* esperar ACK opcional */
    for (volatile int i = 0; i < 10000; i++)
        if (inb(0x64) & 1) { inb(0x60); break; }
    outb(0x64, 0xD4);
    outb(0x60, 0xF4); /* enable */
    mouse_cycle = 0;
}

static void mouse_feed(u8 sc) {
    /* resync: primer byte debe tener bit 3 */
    if (mouse_cycle == 0 && !(sc & 0x08))
        return;
    mouse_pkt[mouse_cycle++] = sc;
    if (mouse_cycle < 3) return;
    mouse_cycle = 0;

    /* overflow bits */
    if (mouse_pkt[0] & 0xC0) return;

    mouse_btn = mouse_pkt[0] & 7;
    int dx = (int)(signed char)mouse_pkt[1];
    int dy = (int)(signed char)mouse_pkt[2];

    /* suavizar: escala /3 y deadzone */
    if (dx > -2 && dx < 2) dx = 0;
    if (dy > -2 && dy < 2) dy = 0;
    dx /= 3;
    dy /= 3;

    int ox = mouse_x, oy = mouse_y;
    mouse_x += dx;
    mouse_y -= dy;
    if (mouse_x < 0) mouse_x = 0;
    if (mouse_y < 0) mouse_y = 0;
    if (fb_ok) {
        if ((u32)mouse_x >= fb_w) mouse_x = (int)fb_w - 1;
        if ((u32)mouse_y >= fb_h) mouse_y = (int)fb_h - 1;
    }

    int mi = buscar("Raton");
    if (mi >= 0) ocbs[mi].disp_dato++;

    if (mouse_btn & 1) {
        if (drag_win < 0) {
            int hit = hit_window(mouse_x, mouse_y);
            if (hit >= 0) {
                focus_set(hit);
                if (mouse_y < ocbs[hit].y + 28) {
                    drag_win = hit;
                    drag_ox = mouse_x - ocbs[hit].x;
                    drag_oy = mouse_y - ocbs[hit].y;
                }
            }
        } else {
            ocbs[drag_win].x = mouse_x - drag_ox;
            ocbs[drag_win].y = mouse_y - drag_oy;
            if (ocbs[drag_win].y < 44) ocbs[drag_win].y = 44;
        }
        mouse_dirty = 1;
    } else {
        drag_win = -1;
        if (ox != mouse_x || oy != mouse_y) mouse_dirty = 1;
    }

    if (mouse_dirty) {
        mouse_dirty = 0;
        compositor_tick();
    }
}

static void text_append(int i, char ch) {
    if (i < 0 || ocbs[i].kind != K_TEXT || ocbs[i].text_len >= MAX_TEXT - 1) return;
    if (igual(ocbs[i].text, "Escribe aqui")) {
        ocbs[i].text[0] = 0;
        ocbs[i].text_len = 0;
    }
    ocbs[i].text[ocbs[i].text_len++] = ch;
    ocbs[i].text[ocbs[i].text_len] = 0;
}
static void text_backspace(int i) {
    if (i < 0 || ocbs[i].kind != K_TEXT || ocbs[i].text_len <= 0) return;
    ocbs[i].text[--ocbs[i].text_len] = 0;
}
static char sc_ascii(u8 sc) {
    static const char map[] =
        "??1234567890-=\b\tqwertyuiop[]\n?asdfghjkl;'`?\\zxcvbnm,./?*? ";
    if (sc >= sizeof(map) - 1) return 0;
    char c = map[sc];
    if (c == '?' || c == '\n' || c == '\t' || c == '\b') return 0;
    return c;
}

static void boot_organisms(void) {
    n_ocb = n_pulso = 0; nonce = 0; focus_win = -1; menu_open = 0; orges_seq = 0;
    ocb_add("Maestro", K_MASTER, CAP_ALL);
    ocb_add("Framebuffer", K_FB, CAP_PULSO_REC | CAP_FB | CAP_DISP_LEER | CAP_DISP_ESCR);
    ocb_add("Compositor", K_COMPOSE, CAP_PULSO_ENV | CAP_PULSO_REC | CAP_COMPOSE | CAP_FB);
    ocb_add("Desktop", K_DESKTOP,
            CAP_PULSO_ENV | CAP_PULSO_REC | CAP_DESKTOP | CAP_ORGES_CREAR | CAP_COMPOSE);
    ocb_add("Consola", K_SHELL, CAP_PULSO_ENV | CAP_PULSO_REC | CAP_ZYRION);
    {
        int c = buscar("Consola");
        ocbs[c].x = 120; ocbs[c].y = 80; ocbs[c].w = 320; ocbs[c].h = 200;
        cpy(ocbs[c].titulo, "Consola", MAX_TITLE);
    }
    ocb_add("BusPulso", K_BUS, CAP_PULSO_ENV | CAP_PULSO_REC);
    ocb_add("Teclado", K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_CTRL);
    ocbs[buscar("Teclado")].disp_clase = 1;
    ocb_add("Raton", K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_CTRL);
    ocbs[buscar("Raton")].disp_clase = 2;
    ocb_add("Disco", K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_ESCR);
    ocbs[buscar("Disco")].disp_clase = 4;
    ocb_add("Red", K_DISP, CAP_PULSO_REC | CAP_DISP_LEER);
    ocbs[buscar("Red")].disp_clase = 5;
    crear_orges(K_WINDOW, "Studio");
    crear_orges(K_TEXT, "Notas");
    crear_orges(K_LIST, "Tareas");
    actor_i = buscar("Desktop");
}

static void vga_msg(const char *s) {
    for (int i = 0; s[i] && i < 80 * 10; i++)
        VGA[i] = (u16)(0x0A << 8) | (u8)s[i];
}
static void vga_desktop(void) {
    for (int i = 0; i < 80 * 25; i++) VGA[i] = (u16)(0x1F << 8) | ' ';
    const char *t = "AlsetOS x64 - VGA texto (sin LFB). Usa -vga std.";
    for (int i = 0; t[i] && i < 78; i++) VGA[i] = (u16)(0x3F << 8) | (u8)t[i];
    int row = 2;
    for (int i = 0; i < n_ocb && row < 22; i++) {
        if (!ocbs[i].alive) continue;
        for (int k = 0; ocbs[i].nombre[k] && k < 20; k++)
            VGA[row * 80 + 2 + k] = (u16)(0x1A << 8) | (u8)ocbs[i].nombre[k];
        row++;
    }
}

void kernel_main(u32 magic, u32 mb_info) {
    nodo_id = fnv("AlsetOS-x64") ^ 0xA15E0064u;
    vga_msg("AlsetOS x86_64 arrancando...");
    mouse_init();
    boot_organisms();
    if (!fb_init(magic, mb_info)) {
        vga_desktop();
        for (;;) {
            if (!(inb(0x64) & 1)) { pausa(); continue; }
            inb(0x60);
            vga_desktop();
        }
    }
    {
        int fi = buscar("Framebuffer");
        if (fi >= 0) { ocbs[fi].w = (int)fb_w; ocbs[fi].h = (int)fb_h; }
    }
    compositor_tick();

    for (;;) {
        if (!(inb(0x64) & 1)) { pausa(); continue; }
        u8 st = inb(0x64);
        u8 sc = inb(0x60);
        if (st & 0x20) { mouse_feed(sc); continue; }
        if (sc == 0xE0 || (sc & 0x80)) continue;

        int ki = buscar("Teclado");
        if (ki >= 0) ocbs[ki].disp_dato++;

        if (sc == 0x32) { menu_open = !menu_open; compositor_tick(); continue; }
        if (sc == 0x02) { crear_orges(K_WINDOW, "Orges"); menu_open = 0; compositor_tick(); continue; }
        if (sc == 0x03) { crear_orges(K_TEXT, "Texto"); menu_open = 0; compositor_tick(); continue; }
        if (sc == 0x04) { crear_orges(K_LIST, "Lista"); menu_open = 0; compositor_tick(); continue; }
        if (sc == 0x1C) {
            int from = buscar("Desktop");
            if (from >= 0 && focus_win >= 0) pulso(from, focus_win);
            compositor_tick(); continue;
        }
        if (sc == 0x0F) {
            int start = focus_win + 1;
            for (int k = 0; k < n_ocb; k++) {
                int i = (start + k) % n_ocb;
                if (ocbs[i].alive && is_win(i)) { focus_set(i); break; }
            }
            compositor_tick(); continue;
        }
        if (sc == 0x0E) { text_backspace(focus_win); compositor_tick(); continue; }
        if (focus_win >= 0 && ocbs[focus_win].kind == K_LIST) {
            if (sc == 0x48 && ocbs[focus_win].sel_item > 0) {
                ocbs[focus_win].sel_item--; compositor_tick(); continue;
            }
            if (sc == 0x50 && ocbs[focus_win].sel_item < ocbs[focus_win].n_items - 1) {
                ocbs[focus_win].sel_item++; compositor_tick(); continue;
            }
        }
        if (focus_win >= 0 && is_win(focus_win)) {
            if (sc == 0x4B) { ocbs[focus_win].x -= 16; compositor_tick(); continue; }
            if (sc == 0x4D) { ocbs[focus_win].x += 16; compositor_tick(); continue; }
            if (sc == 0x48) { ocbs[focus_win].y -= 16; compositor_tick(); continue; }
            if (sc == 0x50) { ocbs[focus_win].y += 16; compositor_tick(); continue; }
        }
        char ch = sc_ascii(sc);
        if (ch && focus_win >= 0 && ocbs[focus_win].kind == K_TEXT) {
            text_append(focus_win, ch);
            compositor_tick();
        }
    }
}
