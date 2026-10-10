/*
 * AlsetOS Genesis — organismos-dispositivo + shell en español + UI gráfica VGA
 * Paradigma: organismo / Pulse / capacidades / Zyrion. Sin Unix.
 */
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;

#define COLS 80
#define ROWS 25
#define MAX_OCB 32
#define MAX_PULSE 64
#define MAX_LINE 76
#define MAX_GOALS 4
#define MAX_GOAL_LEN 14
#define RAMDISK_SECTORS 32
#define RAMDISK_SECSIZE 128

#define ATR_NORM 0x0F
#define ATR_DIM  0x08
#define ATR_ACC  0x0A
#define ATR_TITLE 0x0B
#define ATR_ERR  0x0C
#define ATR_WARN 0x0E
#define ATR_INV  0x1F
#define ATR_SHELL 0x0E

/* Caps */
#define CAP_PULSO_ENV  0x01
#define CAP_PULSO_REC  0x02
#define CAP_ORGES_CREAR 0x04
#define CAP_ORGES_DESTR 0x08
#define CAP_CAPS_OTORG 0x10
#define CAP_NODO_ADMIN 0x20
#define CAP_ZYRION     0x40
#define CAP_DISP_LEER  0x80
#define CAP_DISP_ESCR  0x100
#define CAP_DISP_CTRL  0x200
#define CAP_ALL        0xFFFFFFFFu

/* Tipos de organismo */
#define K_MASTER 1
#define K_SHELL  2
#define K_ORGES  3
#define K_PAR    4
#define K_BUS    5
#define K_DISP   6

static inline u8 inb(u16 p) {
    u8 v; __asm__ __volatile__("inb %1,%0" : "=a"(v) : "Nd"(p)); return v;
}
static inline void outb(u16 p, u8 v) {
    __asm__ __volatile__("outb %0,%1" : : "a"(v), "Nd"(p));
}
static void pausa(void) {
    for (volatile int i = 0; i < 3000; i++)
        __asm__ __volatile__("pause");
}

/* ——— VGA texto ——— */
static volatile u16 *const VGA = (volatile u16*)0xB8000;
static int tx, ty;

static void vput(int x, int y, char c, u8 a) {
    if (x >= 0 && x < COLS && y >= 0 && y < ROWS)
        VGA[y * COLS + x] = (u16)(a << 8) | (u8)c;
}
static void scroll(void) {
    for (int y = 1; y < ROWS - 1; y++)
        for (int x = 0; x < COLS; x++)
            VGA[(y - 1) * COLS + x] = VGA[y * COLS + x];
    for (int x = 0; x < COLS; x++)
        VGA[(ROWS - 2) * COLS + x] = (u16)(ATR_NORM << 8) | ' ';
    if (ty > 0) ty--;
}
static void tclear(void) {
    for (int i = 0; i < COLS * (ROWS - 1); i++)
        VGA[i] = (u16)(ATR_NORM << 8) | ' ';
    tx = ty = 0;
}
static void twrite(const char *s, u8 a) {
    while (*s) {
        if (*s == '\n') {
            tx = 0; ty++;
            if (ty >= ROWS - 1) { scroll(); ty = ROWS - 2; }
            s++; continue;
        }
        vput(tx, ty, *s, a);
        if (++tx >= COLS) { tx = 0; ty++; if (ty >= ROWS - 1) { scroll(); ty = ROWS - 2; } }
        s++;
    }
}
static void tprint(const char *s) { twrite(s, ATR_NORM); }
static void tacc(const char *s) { twrite(s, ATR_ACC); }
static void terr(const char *s) { twrite(s, ATR_ERR); }
static void tdim(const char *s) { twrite(s, ATR_DIM); }
static void thex(u32 v) {
    const char *h = "0123456789ABCDEF";
    char b[11] = "0x";
    for (int i = 0; i < 8; i++) b[2 + i] = h[(v >> (28 - i * 4)) & 0xF];
    b[10] = 0; tacc(b);
}
static void barra_estado(const char *actor, const char *modo) {
    for (int x = 0; x < COLS; x++) vput(x, ROWS - 1, ' ', ATR_INV);
    const char *p = "AlsetOS | ";
    int x = 0;
    for (; p[x]; x++) vput(x, ROWS - 1, p[x], ATR_INV);
    for (int i = 0; modo[i] && x < 20; i++, x++) vput(x, ROWS - 1, modo[i], ATR_INV);
    vput(x++, ROWS - 1, ' ', ATR_INV);
    const char *a = "actor=";
    for (int i = 0; a[i]; i++, x++) vput(x, ROWS - 1, a[i], ATR_INV);
    for (int i = 0; actor[i] && x < COLS - 1; i++, x++) vput(x, ROWS - 1, actor[i], ATR_INV);
}

/* ——— Organismos ——— */
struct OCB {
    u32 magic, oid, kind, caps;
    u32 pin, pout;
    u8 alive;
    char nombre[16];
    char metas[MAX_GOALS][MAX_GOAL_LEN];
    int n_metas;
    /* dispositivo */
    u32 disp_clase; /* 1 teclado 2 raton 3 memoria 4 disco 5 red 6 extraible */
    u32 disp_estado; /* flags runtime */
    u32 disp_dato;   /* contador/bytes */
};

static struct OCB ocbs[MAX_OCB];
static int n_ocb, actor_i, n_pulso;
static u32 nodo_id, nonce;
static char linea[MAX_LINE];
static int lin_len;
static int modo_grafico; /* 0 consola 1 grafico texto-UI */

/* Ramdisk simple (organismo Disco) */
static u8 ramdisk[RAMDISK_SECTORS * RAMDISK_SECSIZE];

/* Ratón PS/2 */
static int mouse_x = 40, mouse_y = 12;
static u8 mouse_btn;
static int mouse_cycle;
static u8 mouse_pkt[3];

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
static int ocb_add(const char *n, u32 kind, u32 caps, u32 dclase) {
    if (n_ocb >= MAX_OCB) return -1;
    if (buscar(n) >= 0) return -2;
    struct OCB *o = &ocbs[n_ocb];
    o->magic = 0x4F434200;
    o->oid = fnv(n) ^ nodo_id ^ (u32)n_ocb;
    o->kind = kind; o->caps = caps;
    o->pin = o->pout = 0; o->alive = 1; o->n_metas = 0;
    o->disp_clase = dclase; o->disp_estado = 1; o->disp_dato = 0;
    cpy(o->nombre, n, 16);
    return n_ocb++;
}

/* Zyrion */
typedef enum { Z_V = 0, Z_F = 1, Z_I = 2 } Z;
static const char *zn(Z z) { return z == Z_V ? "V" : (z == Z_F ? "F" : "I"); }
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
static Z z_not(Z a) { return a == Z_V ? Z_F : (a == Z_F ? Z_V : Z_I); }
static Z z_parse(const char *s) {
    if (igual(s, "V") || igual(s, "v") || igual(s, "verdadero")) return Z_V;
    if (igual(s, "F") || igual(s, "f") || igual(s, "falso")) return Z_F;
    return Z_I;
}

static int pulso(int from, int to, u32 tipo) {
    if (from < 0 || to < 0) return -1;
    if (!tiene_cap(from, CAP_PULSO_ENV)) return -2;
    if (!tiene_cap(to, CAP_PULSO_REC)) return -3;
    ocbs[from].pout++; ocbs[to].pin++;
    nonce++; n_pulso++;
    return 0;
}

static int partir(char *s, char *argv[], int max) {
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

static u32 parse_cap(const char *s) {
    if (igual(s, "pulso.env") || igual(s, "env")) return CAP_PULSO_ENV;
    if (igual(s, "pulso.rec") || igual(s, "rec")) return CAP_PULSO_REC;
    if (igual(s, "orges.crear") || igual(s, "crear")) return CAP_ORGES_CREAR;
    if (igual(s, "orges.destr") || igual(s, "destr")) return CAP_ORGES_DESTR;
    if (igual(s, "caps.otorg") || igual(s, "otorg")) return CAP_CAPS_OTORG;
    if (igual(s, "nodo.admin") || igual(s, "admin")) return CAP_NODO_ADMIN;
    if (igual(s, "zyrion")) return CAP_ZYRION;
    if (igual(s, "disp.leer") || igual(s, "leer")) return CAP_DISP_LEER;
    if (igual(s, "disp.escr") || igual(s, "escr")) return CAP_DISP_ESCR;
    if (igual(s, "disp.ctrl") || igual(s, "ctrl")) return CAP_DISP_CTRL;
    if (igual(s, "todo")) return CAP_ALL;
    return 0;
}

static const char *clase_nombre(u32 c) {
    if (c == 1) return "teclado";
    if (c == 2) return "raton";
    if (c == 3) return "memoria";
    if (c == 4) return "disco";
    if (c == 5) return "red";
    if (c == 6) return "extraible";
    return "-";
}

static void mostrar_org(int i) {
    struct OCB *o = &ocbs[i];
    tprint("  ");
    tacc(o->nombre);
    tprint(" oid="); thex(o->oid);
    tprint(" tipo=");
    if (o->kind == K_MASTER) tprint("MAESTRO");
    else if (o->kind == K_SHELL) tprint("CONSOLA");
    else if (o->kind == K_ORGES) tprint("ORGES");
    else if (o->kind == K_PAR) tprint("PAR");
    else if (o->kind == K_BUS) tprint("BUS");
    else if (o->kind == K_DISP) { tprint("DISP:"); tprint(clase_nombre(o->disp_clase)); }
    else tprint("?");
    tprint(" caps="); thex(o->caps);
    tprint("\n");
}

static void cmd_ayuda(void) {
    tacc("AlsetOS Genesis — shell en español\n");
    tdim("Organismo · Pulso · Capacidades · Zyrion · Dispositivos\n\n");
    tprint("  ayuda                         — esta ayuda\n");
    tprint("  nodo                          — identidad del nodo\n");
    tprint("  organismos | lista            — listar OCB\n");
    tprint("  dispositivos | disp           — solo organismos-dispositivo\n");
    tprint("  actor <nombre>                — cambiar organismo activo\n");
    tprint("  capacidades | caps            — caps del actor\n");
    tprint("  crear orges <nom> [metas..]   — nuevo ORGES\n");
    tprint("  crear par <nombre>            — peer descentralizado\n");
    tprint("  pulso <destino> [tipo]        — enviar Pulso\n");
    tprint("  otorgar <org> <cap>           — Master otorga capacidad\n");
    tprint("  metas <nombre>                — metas de un ORGES\n");
    tprint("  estado <disp>                 — estado de dispositivo\n");
    tprint("  leer <disp> [args]            — leer via organismo-disp\n");
    tprint("  escribir disco <sec> <texto>  — escribir ramdisk\n");
    tprint("  zyrion <A> y|o|no <B>         — logica ternaria\n");
    tprint("  interfaz consola | grafica    — cambiar UI\n");
    tprint("  limpiar                       — limpiar pantalla\n");
    tprint("\nCaps: pulso.env pulso.rec orges.crear orges.destr\n");
    tprint("      caps.otorg nodo.admin zyrion disp.leer disp.escr disp.ctrl\n");
    tprint("Abreviaturas: lista, disp, caps, env, rec, crear, otorg, leer, escr, ctrl\n");
}

static void cmd_estado_disp(int i) {
    struct OCB *o = &ocbs[i];
    if (o->kind != K_DISP) { terr("no es dispositivo\n"); return; }
    tprint("dispositivo "); tacc(o->nombre);
    tprint(" clase="); tprint(clase_nombre(o->disp_clase));
    tprint(" activo="); tprint(o->disp_estado ? "si" : "no");
    tprint(" dato="); thex(o->disp_dato);
    tprint("\n");
    if (o->disp_clase == 3) {
        tprint("  memoria informativa del nodo ( contador de lecturas )\n");
    }
    if (o->disp_clase == 4) {
        tprint("  disco RAM: ");
        thex(RAMDISK_SECTORS);
        tprint(" sectores x ");
        thex(RAMDISK_SECSIZE);
        tprint(" bytes\n");
    }
    if (o->disp_clase == 5) {
        tprint("  red: interfaz local (sin enlace externo en esta fase)\n");
    }
    if (o->disp_clase == 1) {
        tprint("  teclado: eventos="); thex(o->disp_dato); tprint("\n");
    }
    if (o->disp_clase == 2) {
        tprint("  raton x="); thex((u32)mouse_x);
        tprint(" y="); thex((u32)mouse_y);
        tprint(" boton="); thex(mouse_btn); tprint("\n");
    }
}

static void cmd_leer(char *argv[], int argc) {
    if (argc < 2) { terr("uso: leer <dispositivo> [sector]\n"); return; }
    int i = buscar(argv[1]);
    if (i < 0) { terr("no existe\n"); return; }
    if (!tiene_cap(actor_i, CAP_DISP_LEER) && !tiene_cap(actor_i, CAP_DISP_CTRL)) {
        terr("denegado: falta disp.leer\n"); return;
    }
    if (!tiene_cap(i, CAP_DISP_LEER) && ocbs[i].kind == K_DISP) {
        /* el dispositivo debe poder ser leido — caps en el org disp */
    }
    struct OCB *o = &ocbs[i];
    o->disp_dato++;
    if (o->kind != K_DISP) {
        /* leer organismo normal: resumen */
        mostrar_org(i);
        return;
    }
    if (o->disp_clase == 4) {
        u32 sec = 0;
        if (argc >= 3) {
            sec = 0;
            for (const char *p = argv[2]; *p; p++) {
                if (*p >= '0' && *p <= '9') sec = sec * 10 + (u32)(*p - '0');
            }
        }
        if (sec >= RAMDISK_SECTORS) { terr("sector fuera de rango\n"); return; }
        tprint("sector "); thex(sec); tprint(": \"");
        u8 *p = &ramdisk[sec * RAMDISK_SECSIZE];
        for (int k = 0; k < 32 && p[k]; k++) {
            char t[2] = { (char)p[k], 0 };
            if (p[k] >= 32 && p[k] < 127) tprint(t);
            else tprint(".");
        }
        tprint("\"\n");
        pulso(actor_i, i, 10);
        return;
    }
    if (o->disp_clase == 3) {
        tprint("memoria: lecturas="); thex(o->disp_dato);
        tprint("  (heap kernel no expuesto; contador de acceso)\n");
        pulso(actor_i, i, 11);
        return;
    }
    if (o->disp_clase == 5) {
        tprint("red: estado=local  paquetes="); thex(o->disp_dato); tprint("\n");
        pulso(actor_i, i, 12);
        return;
    }
    cmd_estado_disp(i);
    pulso(actor_i, i, 13);
}

static void cmd_escribir(char *argv[], int argc) {
    if (argc < 4 || !igual(argv[1], "disco")) {
        terr("uso: escribir disco <sector> <texto>\n"); return;
    }
    if (!tiene_cap(actor_i, CAP_DISP_ESCR)) {
        terr("denegado: falta disp.escr\n"); return;
    }
    int di = buscar("Disco");
    if (di < 0) { terr("sin organismo Disco\n"); return; }
    u32 sec = 0;
    for (const char *p = argv[2]; *p; p++)
        if (*p >= '0' && *p <= '9') sec = sec * 10 + (u32)(*p - '0');
    if (sec >= RAMDISK_SECTORS) { terr("sector fuera de rango\n"); return; }
    u8 *dst = &ramdisk[sec * RAMDISK_SECSIZE];
    for (int i = 0; i < RAMDISK_SECSIZE; i++) dst[i] = 0;
    const char *txt = argv[3];
    for (int i = 0; i < RAMDISK_SECSIZE - 1 && txt[i]; i++) dst[i] = (u8)txt[i];
    ocbs[di].disp_dato++;
    pulso(actor_i, di, 20);
    tprint("escrito en Disco sector "); thex(sec); tprint("\n");
}

/* ——— Interfaz gráfica (VGA texto enriquecido + ratón) ——— */
static void dibujar_grafica(void) {
    for (int i = 0; i < COLS * ROWS; i++)
        VGA[i] = (u16)(0x1F << 8) | ' ';
    /* marco */
    for (int x = 0; x < COLS; x++) {
        vput(x, 0, ' ', 0x3F);
        vput(x, ROWS - 1, ' ', 0x3F);
    }
    const char *tit = " AlsetOS Genesis  |  Panel de organismos-dispositivo ";
    for (int i = 0; tit[i] && i < COLS; i++) vput(i, 0, tit[i], 0x3F);

    /* columna izquierda: dispositivos */
    vput(0, 1, ' ', 0x1E);
    const char *h1 = " DISPOSITIVOS ";
    for (int i = 0; h1[i]; i++) vput(1 + i, 2, h1[i], 0x1E);
    int row = 3;
    for (int i = 0; i < n_ocb && row < ROWS - 3; i++) {
        if (ocbs[i].kind != K_DISP) continue;
        u8 atr = (i == actor_i) ? 0x2F : 0x1F;
        vput(1, row, (i == actor_i) ? '>' : ' ', atr);
        for (int k = 0; ocbs[i].nombre[k] && k < 14; k++)
            vput(3 + k, row, ocbs[i].nombre[k], atr);
        row++;
    }

    /* panel derecho */
    const char *h2 = " ORGANISMOS / ORGES ";
    for (int i = 0; h2[i]; i++) vput(22 + i, 2, h2[i], 0x1B);
    row = 3;
    for (int i = 0; i < n_ocb && row < 16; i++) {
        if (ocbs[i].kind == K_DISP) continue;
        u8 atr = (i == actor_i) ? 0x2F : 0x0F;
        vput(22, row, (i == actor_i) ? '>' : ' ', atr);
        for (int k = 0; ocbs[i].nombre[k] && k < 12; k++)
            vput(24 + k, row, ocbs[i].nombre[k], atr);
        const char *tp = "?";
        if (ocbs[i].kind == K_MASTER) tp = "MAESTRO";
        else if (ocbs[i].kind == K_SHELL) tp = "CONSOLA";
        else if (ocbs[i].kind == K_ORGES) tp = "ORGES";
        else if (ocbs[i].kind == K_PAR) tp = "PAR";
        else if (ocbs[i].kind == K_BUS) tp = "BUS";
        for (int k = 0; tp[k]; k++) vput(38 + k, row, tp[k], atr);
        row++;
    }

    /* info */
    const char *inf = "Clic o teclas:  Tab cambia foco lista | Enter pulso al seleccionado";
    for (int i = 0; inf[i] && i < COLS - 1; i++) vput(i, ROWS - 3, inf[i], 0x08);
    const char *inf2 = "F1 consola | F2 grafica | 1 crear ORGES demo | W/S mover seleccion";
    for (int i = 0; inf2[i] && i < COLS - 1; i++) vput(i, ROWS - 2, inf2[i], 0x08);

    /* cursor ratón */
    if (mouse_x < 0) mouse_x = 0;
    if (mouse_x >= COLS) mouse_x = COLS - 1;
    if (mouse_y < 0) mouse_y = 0;
    if (mouse_y >= ROWS) mouse_y = ROWS - 1;
    vput(mouse_x, mouse_y, 'X', 0x4F);

    barra_estado(ocbs[actor_i].nombre, "GRAFICA");
}

static void prompt(void) {
    twrite(ocbs[actor_i].nombre, ATR_SHELL);
    twrite(" > ", ATR_SHELL);
}

static void ejecutar(char *buf);

static void graf_seleccion_click(void) {
    /* clic en columna dispositivos (x 1-18) o organismos (x 22+) */
    int row = mouse_y;
    if (row < 3 || row > ROWS - 4) return;
    int idx_disp = 0;
    int idx_org = 0;
    for (int i = 0; i < n_ocb; i++) {
        if (ocbs[i].kind == K_DISP) {
            if (3 + idx_disp == row && mouse_x < 20) {
                actor_i = i;
                dibujar_grafica();
                return;
            }
            idx_disp++;
        } else {
            if (3 + idx_org == row && mouse_x >= 20) {
                actor_i = i;
                dibujar_grafica();
                return;
            }
            idx_org++;
        }
    }
}

static void ejecutar(char *buf) {
    char *argv[14];
    int argc = partir(buf, argv, 14);
    if (argc == 0) return;

    if (igual(argv[0], "ayuda") || igual(argv[0], "aid") || igual(argv[0], "?")) {
        cmd_ayuda(); return;
    }
    if (igual(argv[0], "limpiar") || igual(argv[0], "cls")) {
        tclear(); return;
    }
    if (igual(argv[0], "nodo")) {
        tprint("NodoID "); thex(nodo_id);
        tprint("  actor="); tacc(ocbs[actor_i].nombre); tprint("\n");
        return;
    }
    if (igual(argv[0], "organismos") || igual(argv[0], "lista") || igual(argv[0], "ls")) {
        tprint("Organismos del nodo:\n");
        for (int i = 0; i < n_ocb; i++) if (ocbs[i].alive) mostrar_org(i);
        return;
    }
    if (igual(argv[0], "dispositivos") || igual(argv[0], "disp")) {
        tprint("Organismos-dispositivo:\n");
        for (int i = 0; i < n_ocb; i++)
            if (ocbs[i].alive && ocbs[i].kind == K_DISP) {
                mostrar_org(i);
                tdim("    clase="); tdim(clase_nombre(ocbs[i].disp_clase));
                tprint(" estado="); tprint(ocbs[i].disp_estado ? "activo" : "inactivo");
                tprint("\n");
            }
        return;
    }
    if (igual(argv[0], "actor") && argc >= 2) {
        int i = buscar(argv[1]);
        if (i < 0) { terr("organismo desconocido\n"); return; }
        actor_i = i;
        tprint("actor = "); tacc(ocbs[i].nombre); tprint("\n");
        return;
    }
    if (igual(argv[0], "capacidades") || igual(argv[0], "caps")) {
        tprint("caps de "); tacc(ocbs[actor_i].nombre);
        tprint(": "); thex(ocbs[actor_i].caps); tprint("\n");
        return;
    }
    if (igual(argv[0], "crear") && argc >= 3) {
        if (!tiene_cap(actor_i, CAP_ORGES_CREAR)) {
            terr("denegado: falta orges.crear\n"); return;
        }
        u32 kind = K_ORGES; u32 caps = CAP_PULSO_REC | CAP_PULSO_ENV;
        if (igual(argv[1], "par")) { kind = K_PAR; }
        else if (!igual(argv[1], "orges")) {
            terr("uso: crear orges|par <nombre> [metas..]\n"); return;
        }
        int id = ocb_add(argv[2], kind, caps, 0);
        if (id == -2) { terr("nombre duplicado\n"); return; }
        if (id < 0) { terr("tabla llena\n"); return; }
        for (int a = 3; a < argc && ocbs[id].n_metas < MAX_GOALS; a++) {
            cpy(ocbs[id].metas[ocbs[id].n_metas], argv[a], MAX_GOAL_LEN);
            ocbs[id].n_metas++;
        }
        tprint("creado "); tacc(argv[2]); tprint(" oid="); thex(ocbs[id].oid); tprint("\n");
        return;
    }
    if (igual(argv[0], "pulso") && argc >= 2) {
        int to = buscar(argv[1]);
        if (to < 0) { terr("destino desconocido\n"); return; }
        int r = pulso(actor_i, to, 1);
        if (r == -2) { terr("denegado: falta pulso.env\n"); return; }
        if (r == -3) { terr("denegado: destino sin pulso.rec\n"); return; }
        tprint("Pulso "); tacc(ocbs[actor_i].nombre); tprint(" -> ");
        tacc(ocbs[to].nombre); tprint(" nonce="); thex(nonce); tprint("\n");
        return;
    }
    if (igual(argv[0], "otorgar") && argc >= 3) {
        if (!tiene_cap(actor_i, CAP_CAPS_OTORG) && ocbs[actor_i].kind != K_MASTER) {
            terr("denegado: falta caps.otorg\n"); return;
        }
        int t = buscar(argv[1]);
        if (t < 0) { terr("no existe\n"); return; }
        u32 c = parse_cap(argv[2]);
        if (!c) { terr("capacidad desconocida\n"); return; }
        ocbs[t].caps |= c;
        tprint("otorgado "); tprint(argv[2]); tprint(" a "); tacc(argv[1]); tprint("\n");
        return;
    }
    if (igual(argv[0], "metas") && argc >= 2) {
        int i = buscar(argv[1]);
        if (i < 0) { terr("no existe\n"); return; }
        if (!ocbs[i].n_metas) { tdim("(sin metas)\n"); return; }
        for (int g = 0; g < ocbs[i].n_metas; g++) {
            tprint("  - "); tprint(ocbs[i].metas[g]); tprint("\n");
        }
        return;
    }
    if (igual(argv[0], "estado") && argc >= 2) {
        int i = buscar(argv[1]);
        if (i < 0) { terr("no existe\n"); return; }
        cmd_estado_disp(i);
        return;
    }
    if (igual(argv[0], "leer")) { cmd_leer(argv, argc); return; }
    if (igual(argv[0], "escribir")) { cmd_escribir(argv, argc); return; }
    if (igual(argv[0], "zyrion") && argc >= 2) {
        if (!tiene_cap(actor_i, CAP_ZYRION) && ocbs[actor_i].kind != K_MASTER) {
            terr("denegado: falta zyrion\n"); return;
        }
        if (argc >= 3 && igual(argv[1], "no")) {
            Z a = z_parse(argv[2]);
            tprint("no "); tprint(zn(a)); tprint(" = "); tacc(zn(z_not(a))); tprint("\n");
            return;
        }
        if (argc >= 4) {
            Z a = z_parse(argv[1]); Z b = z_parse(argv[3]); Z r = Z_I;
            if (igual(argv[2], "y")) r = z_and(a, b);
            else if (igual(argv[2], "o")) r = z_or(a, b);
            else { terr("uso: zyrion V y I | zyrion no I\n"); return; }
            tprint(zn(a)); tprint(" "); tprint(argv[2]); tprint(" ");
            tprint(zn(b)); tprint(" = "); tacc(zn(r)); tprint("\n");
            return;
        }
        terr("uso: zyrion V y I\n"); return;
    }
    if (igual(argv[0], "interfaz") && argc >= 2) {
        if (igual(argv[1], "grafica") || igual(argv[1], "gui")) {
            modo_grafico = 1;
            dibujar_grafica();
            return;
        }
        if (igual(argv[1], "consola") || igual(argv[1], "cli")) {
            modo_grafico = 0;
            tclear();
            tacc("Consola AlsetOS\n");
            prompt();
            return;
        }
    }
    terr("comando desconocido (ayuda)\n");
}

/* ——— PS/2 teclado + ratón ——— */
static void kbd_init(void) {
    for (int i = 0; i < 256; i++) {
        if (!(inb(0x64) & 1)) break;
        (void)inb(0x60);
    }
    outb(0x64, 0xAE); /* kbd enable */
    /* intentar habilitar ratón auxiliar */
    outb(0x64, 0xA8);
    outb(0x64, 0xD4);
    outb(0x60, 0xF4);
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

static void mouse_feed(u8 b) {
    mouse_pkt[mouse_cycle++] = b;
    if (mouse_cycle < 3) return;
    mouse_cycle = 0;
    if (!(mouse_pkt[0] & 0x08)) return; /* resync */
    mouse_btn = mouse_pkt[0] & 0x07;
    int dx = (int)(signed char)mouse_pkt[1];
    int dy = (int)(signed char)mouse_pkt[2];
    mouse_x += dx / 2;
    mouse_y -= dy / 2;
    if (mouse_x < 0) mouse_x = 0;
    if (mouse_x >= COLS) mouse_x = COLS - 1;
    if (mouse_y < 0) mouse_y = 0;
    if (mouse_y >= ROWS) mouse_y = ROWS - 1;
    int mi = buscar("Raton");
    if (mi >= 0) ocbs[mi].disp_dato++;
    if (modo_grafico) {
        dibujar_grafica();
        if (mouse_btn & 1)
            graf_seleccion_click();
    }
}

static void boot(void) {
    nodo_id = fnv("AlsetGenesis") ^ 0xA15E0002u;
    n_ocb = n_pulso = 0; nonce = 0; modo_grafico = 0;
    for (int i = 0; i < (int)sizeof(ramdisk); i++) ramdisk[i] = 0;

    ocb_add("Maestro",  K_MASTER, CAP_ALL, 0);
    ocb_add("Consola",  K_SHELL,
            CAP_PULSO_ENV | CAP_PULSO_REC | CAP_ZYRION | CAP_ORGES_CREAR |
            CAP_DISP_LEER | CAP_DISP_ESCR | CAP_DISP_CTRL, 0);
    ocb_add("BusPulso", K_BUS, CAP_PULSO_ENV | CAP_PULSO_REC, 0);

    /* organismos-dispositivo: el hardware se modela como organismo */
    ocb_add("Teclado",   K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_CTRL, 1);
    ocb_add("Raton",     K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_CTRL, 2);
    ocb_add("Memoria",   K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_CTRL, 3);
    ocb_add("Disco",     K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_ESCR | CAP_DISP_CTRL, 4);
    ocb_add("Red",       K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_CTRL, 5);
    ocb_add("Extraible", K_DISP, CAP_PULSO_REC | CAP_DISP_LEER | CAP_DISP_CTRL, 6);

    actor_i = buscar("Consola");
}

void kernel_main(u32 magic, u32 mb_info) {
    (void)magic; (void)mb_info;
    kbd_init();
    boot();
    tclear();
    tacc("AlsetOS Genesis\n");
    tdim("Organismos-dispositivo · Pulso · Capacidades · UI consola/grafica\n\n");
    tprint("NodoID "); thex(nodo_id); tprint("\n");
    tprint("Escribe "); tacc("ayuda"); tprint(" o "); tacc("interfaz grafica"); tprint("\n\n");
    barra_estado(ocbs[actor_i].nombre, "CONSOLA");
    prompt();
    lin_len = 0;

    for (;;) {
        if (!(inb(0x64) & 1)) { pausa(); continue; }
        u8 st = inb(0x64);
        u8 sc = inb(0x60);

        /* ratón: bit 5 de status indica origen auxiliar en muchos controladores */
        if (st & 0x20) {
            mouse_feed(sc);
            continue;
        }

        if (sc == 0xE0) continue;
        if (sc & 0x80) continue;

        int ki = buscar("Teclado");
        if (ki >= 0) ocbs[ki].disp_dato++;

        /* F1 = consola, F2 = grafica */
        if (sc == 0x3B) { /* F1 */
            modo_grafico = 0;
            tclear();
            tacc("Consola\n");
            prompt();
            continue;
        }
        if (sc == 0x3C) { /* F2 */
            modo_grafico = 1;
            dibujar_grafica();
            continue;
        }

        if (modo_grafico) {
            if (sc == 0x11 || sc == 0x48) { /* W / up */
                if (actor_i > 0) actor_i--;
                dibujar_grafica();
            } else if (sc == 0x1F || sc == 0x50) {
                if (actor_i < n_ocb - 1) actor_i++;
                dibujar_grafica();
            } else if (sc == 0x1C || sc == 0x39) {
                /* pulso Consola -> seleccionado */
                int from = buscar("Consola");
                if (from >= 0) pulso(from, actor_i, 1);
                dibujar_grafica();
            } else if (sc == 0x02) {
                /* 1: crear ORGES demo */
                int from = buscar("Consola");
                if (from >= 0 && tiene_cap(from, CAP_ORGES_CREAR)) {
                    actor_i = from;
                    ocb_add("Demo", K_ORGES, CAP_PULSO_ENV | CAP_PULSO_REC, 0);
                }
                dibujar_grafica();
            }
            continue;
        }

        /* modo consola: línea de comandos */
        if (sc == 0x0E) {
            if (lin_len > 0) {
                lin_len--; linea[lin_len] = 0;
                if (tx > 0) { tx--; vput(tx, ty, ' ', ATR_NORM); }
            }
            continue;
        }
        char ch = sc_ascii(sc);
        if (!ch) continue;
        if (ch == '\n') {
            tprint("\n");
            linea[lin_len] = 0;
            if (lin_len > 0) ejecutar(linea);
            lin_len = 0;
            if (!modo_grafico) {
                barra_estado(ocbs[actor_i].nombre, "CONSOLA");
                prompt();
            }
            continue;
        }
        if (lin_len < MAX_LINE - 1 && ch >= 32) {
            linea[lin_len++] = ch;
            char t[2] = { ch, 0 };
            twrite(t, ATR_NORM);
        }
    }
}
