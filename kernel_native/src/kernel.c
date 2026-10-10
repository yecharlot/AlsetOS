/*
 * AlsetOS Genesis Kernel — fiel al paradigma ATS (organismo, no proceso)
 * Sin Unix: no hay procesos, usuarios POSIX ni filesystem como unidad primaria.
 * Unidad = Organismo (OCB). Interacción = Pulse. Autoridad = Capabilities.
 * Shell declarativa orientada a ORGES. Zyrion V|F|I. Default-deny.
 */
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long long u64;

#define VGA ((volatile u16*)0xB8000)
#define COLS 80
#define ROWS 25
#define MAX_OCB 24
#define MAX_PULSE 64
#define MAX_LINE 72
#define MAX_GOALS 4
#define MAX_GOAL_LEN 12
#define MAX_CAPS 8

#define ATR_NORM 0x0F
#define ATR_DIM  0x08
#define ATR_ACC  0x0A
#define ATR_TITLE 0x0B
#define ATR_WARN 0x0E
#define ATR_ERR  0x0C
#define ATR_INV  0x1F
#define ATR_SHELL 0x0E

/* ——— I/O ——— */
static inline u8 inb(u16 p) {
    u8 v; __asm__ __volatile__("inb %1,%0" : "=a"(v) : "Nd"(p)); return v;
}
static inline void outb(u16 p, u8 v) {
    __asm__ __volatile__("outb %0,%1" : : "a"(v), "Nd"(p));
}
static void pause_poll(void) {
    for (volatile int i = 0; i < 4000; i++)
        __asm__ __volatile__("pause");
}

/* ——— VGA scroll terminal ——— */
static int term_y = 0;
static u8 term_attr = ATR_NORM;

static void vgaput(int x, int y, char c, u8 a) {
    if (x >= 0 && x < COLS && y >= 0 && y < ROWS)
        VGA[y * COLS + x] = (u16)(a << 8) | (u8)c;
}

static void scroll_up(void) {
    for (int y = 1; y < ROWS - 1; y++)
        for (int x = 0; x < COLS; x++)
            VGA[(y - 1) * COLS + x] = VGA[y * COLS + x];
    for (int x = 0; x < COLS; x++)
        VGA[(ROWS - 2) * COLS + x] = (u16)(ATR_NORM << 8) | ' ';
    if (term_y > 0) term_y--;
}

static void tputc(char c) {
    if (c == '\n') {
        term_y++;
        if (term_y >= ROWS - 1) { scroll_up(); term_y = ROWS - 2; }
        return;
    }
    /* find end of current line content approx via writing left to right */
}

static int term_x = 0;

static void tclear(void) {
    for (int i = 0; i < COLS * (ROWS - 1); i++)
        VGA[i] = (u16)(ATR_NORM << 8) | ' ';
    term_x = 0;
    term_y = 0;
}

static void twrite(const char *s, u8 a) {
    while (*s) {
        if (*s == '\n') {
            term_x = 0;
            term_y++;
            if (term_y >= ROWS - 1) { scroll_up(); term_y = ROWS - 2; }
            s++;
            continue;
        }
        vgaput(term_x, term_y, *s, a);
        term_x++;
        if (term_x >= COLS) {
            term_x = 0;
            term_y++;
            if (term_y >= ROWS - 1) { scroll_up(); term_y = ROWS - 2; }
        }
        s++;
    }
}

static void tprint(const char *s) { twrite(s, ATR_NORM); }
static void tacc(const char *s) { twrite(s, ATR_ACC); }
static void terr(const char *s) { twrite(s, ATR_ERR); }
static void twarn(const char *s) { twrite(s, ATR_WARN); }
static void tdim(const char *s) { twrite(s, ATR_DIM); }

static void thex(u32 v) {
    const char *h = "0123456789ABCDEF";
    char b[11];
    b[0] = '0'; b[1] = 'x';
    for (int i = 0; i < 8; i++) b[2 + i] = h[(v >> (28 - i * 4)) & 0xF];
    b[10] = 0;
    tacc(b);
}

static void draw_status(const char *actor) {
    for (int x = 0; x < COLS; x++)
        vgaput(x, ROWS - 1, ' ', ATR_INV);
    const char *p = "AlsetOS Genesis | actor=";
    int x = 0;
    for (; p[x]; x++) vgaput(x, ROWS - 1, p[x], ATR_INV);
    for (int i = 0; actor[i] && x < COLS - 1; i++, x++)
        vgaput(x, ROWS - 1, actor[i], ATR_INV);
}

/* ——— Seguridad: capacidades (default-deny) ——— */
/* Cap IDs (bits / tokens) */
#define CAP_PULSE_SEND    0x01
#define CAP_PULSE_RECV    0x02
#define CAP_ORGES_CREATE  0x04
#define CAP_ORGES_DESTROY 0x08
#define CAP_CAPS_GRANT    0x10
#define CAP_NODE_ADMIN    0x20
#define CAP_ZYRION_EVAL   0x40
#define CAP_ALL           0xFFFFFFFFu

struct OCB {
    u32 magic;
    u32 oid;
    u32 kind; /* 1 master 2 shell 3 orges 4 peer 5 bus */
    u32 caps;
    u32 pulse_in, pulse_out;
    u8  alive;
    char name[16];
    char goals[MAX_GOALS][MAX_GOAL_LEN];
    int n_goals;
};

struct PulseMsg {
    u32 from_oid, to_oid, type, nonce;
    u32 ok; /* 1 delivered */
};

static struct OCB ocbs[MAX_OCB];
static int n_ocb;
static struct PulseMsg logp[MAX_PULSE];
static int n_logp;
static u32 node_id;
static u32 nonce_ctr;
static int actor_idx; /* organismo que “habla” en la shell */
static char line[MAX_LINE];
static int line_len;

static u32 fnv(const char *s) {
    u32 h = 2166136261u;
    while (*s) { h ^= (u8)*s++; h *= 16777619u; }
    return h;
}

static void cpy(char *d, const char *s, int n) {
    int i = 0;
    for (; i < n - 1 && s[i]; i++) d[i] = s[i];
    d[i] = 0;
}

static int streq(const char *a, const char *b) {
    while (*a && *b && *a == *b) { a++; b++; }
    return *a == 0 && *b == 0;
}

static int str_starts(const char *s, const char *p) {
    while (*p) { if (*s++ != *p++) return 0; }
    return 1;
}

static int find_org(const char *name) {
    for (int i = 0; i < n_ocb; i++)
        if (ocbs[i].alive && streq(ocbs[i].name, name)) return i;
    return -1;
}

static int has_cap(int idx, u32 cap) {
    if (idx < 0 || idx >= n_ocb || !ocbs[idx].alive) return 0;
    if (ocbs[idx].kind == 1) return 1; /* Master soberano del nodo */
    return (ocbs[idx].caps & cap) != 0;
}

static int ocb_add(const char *name, u32 kind, u32 caps) {
    if (n_ocb >= MAX_OCB) return -1;
    if (find_org(name) >= 0) return -2;
    struct OCB *o = &ocbs[n_ocb];
    o->magic = 0x4F434200;
    o->oid = fnv(name) ^ node_id ^ (u32)n_ocb;
    o->kind = kind;
    o->caps = caps;
    o->pulse_in = o->pulse_out = 0;
    o->alive = 1;
    o->n_goals = 0;
    cpy(o->name, name, 16);
    return n_ocb++;
}

/* Zyrion ternario: V=0 F=1 I=2 */
typedef enum { Z_V = 0, Z_F = 1, Z_I = 2 } Z;
static const char *zname(Z z) {
    return z == Z_V ? "V" : (z == Z_F ? "F" : "I");
}
static Z z_and(Z a, Z b) {
    if (a == Z_F || b == Z_F) return Z_F;
    if (a == Z_I || b == Z_I) return Z_I;
    return Z_V;
}
static Z z_or(Z a, Z b) {
    if (a == Z_V || b == Z_V) return Z_V;
    if (a == Z_I || b == Z_I) return Z_I;
    return Z_F;
}
static Z z_not(Z a) {
    if (a == Z_V) return Z_F;
    if (a == Z_F) return Z_V;
    return Z_I;
}
static Z z_parse(const char *s) {
    if (streq(s, "V") || streq(s, "v") || streq(s, "true")) return Z_V;
    if (streq(s, "F") || streq(s, "f") || streq(s, "false")) return Z_F;
    return Z_I;
}

/* ——— Pulse con chequeo de capacidades ——— */
static int pulse_do(int from, int to, u32 type) {
    if (from < 0 || to < 0) return -1;
    if (!has_cap(from, CAP_PULSE_SEND)) return -2;
    if (!has_cap(to, CAP_PULSE_RECV)) return -3;
    if (n_logp >= MAX_PULSE) n_logp = 0;
    logp[n_logp].from_oid = ocbs[from].oid;
    logp[n_logp].to_oid = ocbs[to].oid;
    logp[n_logp].type = type;
    logp[n_logp].nonce = ++nonce_ctr;
    logp[n_logp].ok = 1;
    ocbs[from].pulse_out++;
    ocbs[to].pulse_in++;
    n_logp++;
    return 0;
}

/* ——— Tokenizer simple ——— */
static int split(char *s, char *argv[], int max) {
    int n = 0, i = 0;
    while (s[i] && n < max) {
        while (s[i] == ' ' || s[i] == '\t') i++;
        if (!s[i]) break;
        argv[n++] = &s[i];
        while (s[i] && s[i] != ' ' && s[i] != '\t') i++;
        if (s[i]) s[i++] = 0;
    }
    return n;
}

static void show_org(int i) {
    struct OCB *o = &ocbs[i];
    tprint("  ");
    tacc(o->name);
    tprint("  oid=");
    thex(o->oid);
    tprint("  kind=");
    if (o->kind == 1) tprint("MASTER");
    else if (o->kind == 2) tprint("SHELL");
    else if (o->kind == 3) tprint("ORGES");
    else if (o->kind == 4) tprint("PEER");
    else if (o->kind == 5) tprint("BUS");
    else tprint("?");
    tprint("  caps=");
    thex(o->caps);
    tprint("  pin=");
    thex(o->pulse_in);
    tprint(" pout=");
    thex(o->pulse_out);
    if (o->n_goals) {
        tprint("\n    goals:");
        for (int g = 0; g < o->n_goals; g++) {
            tprint(" ");
            tdim(o->goals[g]);
        }
    }
    tprint("\n");
}

static void cmd_help(void) {
    tacc("AlsetOS Genesis — shell declarativa de organismos\n");
    tdim("No hay procesos Unix. Unidad = organismo. Via = Pulse. Seguridad = caps.\n\n");
    tprint("  help\n");
    tprint("  node                         — identidad del nodo\n");
    tprint("  organisms | ls               — listar OCB\n");
    tprint("  actor <nombre>               — organismo activo (quien ejecuta)\n");
    tprint("  create orges <nombre> [goal..] — crear ORGES (requiere CAP)\n");
    tprint("  create peer <nombre>         — peer descentralizado\n");
    tprint("  pulse <destino> [tipo]       — Pulse actor -> destino\n");
    tprint("  grant <nombre> <cap>         — Master otorga capacidad\n");
    tprint("  caps                         — capacidades del actor\n");
    tprint("  zyrion <A> and|or|not <B>    — logica ternaria V F I\n");
    tprint("  goals <nombre>               — ver goals de un ORGES\n");
    tprint("  clear\n");
    tprint("\nCaps: pulse.send pulse.recv orges.create orges.destroy\n");
    tprint("      caps.grant node.admin zyrion.eval\n");
    tprint("Default-deny: sin cap, la accion falla.\n");
}

static u32 parse_cap(const char *s) {
    if (streq(s, "pulse.send") || streq(s, "send")) return CAP_PULSE_SEND;
    if (streq(s, "pulse.recv") || streq(s, "recv")) return CAP_PULSE_RECV;
    if (streq(s, "orges.create") || streq(s, "create")) return CAP_ORGES_CREATE;
    if (streq(s, "orges.destroy") || streq(s, "destroy")) return CAP_ORGES_DESTROY;
    if (streq(s, "caps.grant") || streq(s, "grant")) return CAP_CAPS_GRANT;
    if (streq(s, "node.admin") || streq(s, "admin")) return CAP_NODE_ADMIN;
    if (streq(s, "zyrion.eval") || streq(s, "zyrion")) return CAP_ZYRION_EVAL;
    if (streq(s, "all")) return CAP_ALL;
    return 0;
}

static void exec_line(char *buf) {
    char *argv[12];
    int argc = split(buf, argv, 12);
    if (argc == 0) return;

    if (streq(argv[0], "help") || streq(argv[0], "?")) {
        cmd_help();
        return;
    }
    if (streq(argv[0], "clear")) {
        tclear();
        return;
    }
    if (streq(argv[0], "node")) {
        tprint("NodeID ");
        thex(node_id);
        tprint("  (soberano local, sin registro central)\n");
        tprint("Actor ");
        tacc(ocbs[actor_idx].name);
        tprint("\n");
        return;
    }
    if (streq(argv[0], "organisms") || streq(argv[0], "ls")) {
        tprint("Organismos del nodo:\n");
        for (int i = 0; i < n_ocb; i++)
            if (ocbs[i].alive) show_org(i);
        return;
    }
    if (streq(argv[0], "actor") && argc >= 2) {
        int i = find_org(argv[1]);
        if (i < 0) { terr("organismo desconocido\n"); return; }
        actor_idx = i;
        tprint("actor = ");
        tacc(ocbs[i].name);
        tprint("\n");
        draw_status(ocbs[actor_idx].name);
        return;
    }
    if (streq(argv[0], "caps")) {
        tprint("caps de ");
        tacc(ocbs[actor_idx].name);
        tprint(": ");
        thex(ocbs[actor_idx].caps);
        tprint("\n");
        return;
    }
    if (streq(argv[0], "goals") && argc >= 2) {
        int i = find_org(argv[1]);
        if (i < 0) { terr("no existe\n"); return; }
        if (!ocbs[i].n_goals) { tdim("(sin goals)\n"); return; }
        for (int g = 0; g < ocbs[i].n_goals; g++) {
            tprint("  - ");
            tprint(ocbs[i].goals[g]);
            tprint("\n");
        }
        return;
    }
    if (streq(argv[0], "create") && argc >= 3) {
        if (!has_cap(actor_idx, CAP_ORGES_CREATE) && !has_cap(actor_idx, CAP_NODE_ADMIN)) {
            terr("denegado: falta orges.create\n");
            return;
        }
        u32 kind = 3;
        u32 caps = CAP_PULSE_RECV;
        if (streq(argv[1], "peer")) {
            kind = 4;
            caps = CAP_PULSE_RECV | CAP_PULSE_SEND;
        } else if (streq(argv[1], "orges")) {
            kind = 3;
            caps = CAP_PULSE_RECV | CAP_PULSE_SEND;
        } else {
            terr("uso: create orges|peer <nombre> [goals...]\n");
            return;
        }
        int id = ocb_add(argv[2], kind, caps);
        if (id == -2) { terr("nombre duplicado\n"); return; }
        if (id < 0) { terr("tabla OCB llena\n"); return; }
        for (int a = 3; a < argc && ocbs[id].n_goals < MAX_GOALS; a++) {
            cpy(ocbs[id].goals[ocbs[id].n_goals], argv[a], MAX_GOAL_LEN);
            ocbs[id].n_goals++;
        }
        tprint("creado ");
        tacc(argv[2]);
        tprint(" oid=");
        thex(ocbs[id].oid);
        tprint("\n");
        return;
    }
    if (streq(argv[0], "pulse") && argc >= 2) {
        int to = find_org(argv[1]);
        if (to < 0) { terr("destino desconocido\n"); return; }
        u32 typ = 1;
        if (argc >= 3) typ = fnv(argv[2]) & 0xFF;
        int r = pulse_do(actor_idx, to, typ);
        if (r == -2) { terr("denegado: actor sin pulse.send\n"); return; }
        if (r == -3) { terr("denegado: destino sin pulse.recv\n"); return; }
        if (r != 0) { terr("error pulse\n"); return; }
        tprint("Pulse ");
        tacc(ocbs[actor_idx].name);
        tprint(" -> ");
        tacc(ocbs[to].name);
        tprint(" nonce=");
        thex(nonce_ctr);
        tprint("\n");
        return;
    }
    if (streq(argv[0], "grant") && argc >= 3) {
        if (!has_cap(actor_idx, CAP_CAPS_GRANT) && ocbs[actor_idx].kind != 1) {
            terr("denegado: falta caps.grant\n");
            return;
        }
        int t = find_org(argv[1]);
        if (t < 0) { terr("organismo desconocido\n"); return; }
        u32 c = parse_cap(argv[2]);
        if (!c) { terr("cap desconocida\n"); return; }
        ocbs[t].caps |= c;
        tprint("otorgado ");
        tprint(argv[2]);
        tprint(" a ");
        tacc(argv[1]);
        tprint("\n");
        return;
    }
    if (streq(argv[0], "zyrion") && argc >= 2) {
        if (!has_cap(actor_idx, CAP_ZYRION_EVAL) && ocbs[actor_idx].kind != 1) {
            terr("denegado: falta zyrion.eval\n");
            return;
        }
        if (argc == 3 && streq(argv[1], "not")) {
            Z a = z_parse(argv[2]);
            tprint("not ");
            tprint(zname(a));
            tprint(" = ");
            tacc(zname(z_not(a)));
            tprint("\n");
            return;
        }
        if (argc >= 4) {
            Z a = z_parse(argv[1]);
            Z b = z_parse(argv[3]);
            Z r = Z_I;
            if (streq(argv[2], "and")) r = z_and(a, b);
            else if (streq(argv[2], "or")) r = z_or(a, b);
            else { terr("uso: zyrion A and|or B | zyrion not A\n"); return; }
            tprint(zname(a));
            tprint(" ");
            tprint(argv[2]);
            tprint(" ");
            tprint(zname(b));
            tprint(" = ");
            tacc(zname(r));
            tprint("\n");
            return;
        }
        terr("uso: zyrion V and I\n");
        return;
    }
    terr("comando desconocido (help)\n");
}

static void prompt(void) {
    twrite(ocbs[actor_idx].name, ATR_SHELL);
    twrite(" > ", ATR_SHELL);
}

static void kbd_init(void) {
    for (int i = 0; i < 256; i++) {
        if (!(inb(0x64) & 1)) break;
        (void)inb(0x60);
    }
    outb(0x64, 0xAE);
}

static char sc_ascii(u8 sc) {
    static const char map[64] = {
        0,0,'1','2','3','4','5','6','7','8','9','0','-','=',0,'\t',
        'q','w','e','r','t','y','u','i','o','p','[',']','\n',0,
        'a','s','d','f','g','h','j','k','l',';','\'',0,0,'\\',
        'z','x','c','v','b','n','m',',','.','/',0,0,0,' '
    };
    if (sc >= 64) return 0;
    return map[sc];
}

static void boot_genesis(void) {
    node_id = fnv("AlsetGenesis") ^ 0xA15E0001u;
    n_ocb = 0;
    n_logp = 0;
    nonce_ctr = 0;
    /* Master: soberano del nodo — no es root Unix */
    ocb_add("Master", 1, CAP_ALL);
    /* Shell: organismo de interaccion humana */
    ocb_add("Shell", 2, CAP_PULSE_SEND | CAP_PULSE_RECV | CAP_ZYRION_EVAL | CAP_ORGES_CREATE);
    /* Bus de Pulse: infraestructura */
    ocb_add("PulseBus", 5, CAP_PULSE_SEND | CAP_PULSE_RECV);
    actor_idx = 1; /* Shell por defecto (no Master, least privilege de consola) */
}

void kernel_main(u32 magic, u32 mb_info) {
    (void)magic;
    (void)mb_info;
    kbd_init();
    boot_genesis();
    tclear();
    tacc("AlsetOS Genesis Kernel\n");
    tdim("Organismo · Pulse · Capabilities · Zyrion · ORGES\n");
    tdim("Default-deny · Node soberano · sin Unix\n\n");
    tprint("NodeID ");
    thex(node_id);
    tprint("\nEscribe ");
    tacc("help");
    tprint(" para comenzar.\n\n");
    draw_status(ocbs[actor_idx].name);
    prompt();

    line_len = 0;
    int ext = 0;
    for (;;) {
        if (!(inb(0x64) & 1)) {
            pause_poll();
            continue;
        }
        u8 sc = inb(0x60);
        if (sc == 0xE0) { ext = 1; continue; }
        if (sc & 0x80) { ext = 0; continue; }

        /* backspace */
        if (sc == 0x0E) {
            if (line_len > 0) {
                line_len--;
                line[line_len] = 0;
                if (term_x > 0) {
                    term_x--;
                    vgaput(term_x, term_y, ' ', ATR_NORM);
                }
            }
            ext = 0;
            continue;
        }
        char ch = sc_ascii(sc);
        if (!ch) { ext = 0; continue; }
        if (ch == '\n') {
            tprint("\n");
            line[line_len] = 0;
            if (line_len > 0)
                exec_line(line);
            line_len = 0;
            draw_status(ocbs[actor_idx].name);
            prompt();
            ext = 0;
            continue;
        }
        if (line_len < MAX_LINE - 1 && ch >= 32) {
            line[line_len++] = ch;
            char tmp[2] = { ch, 0 };
            twrite(tmp, ATR_NORM);
        }
        ext = 0;
        (void)ext;
    }
}
