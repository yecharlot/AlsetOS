/*
 * AlsetOS kernel_native — Desktop (clon visual del shell alsetos_native)
 *
 * MAPA DE APRENDIZAJE (lee docs/LEARN_DESKTOP.md):
 *   desktop/shell/shell.css  → colores, topbar/taskbar, ventanas, menú
 *   desktop/shell/shell.js   → foco, z-order, arrastre, iconos, taskbar
 *   cmd/alsetos-native       → núcleo ATS-001 + UI embebida (Go/HTTP)
 *   ESTE ARCHIVO             → mismos conceptos en framebuffer bare-metal
 *
 * Todo sigue siendo organismo (OCB). No imita Windows: clona Alset Desktop.
 */
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long u64;
typedef signed char i8;

#define MAX_OCB 64
#define MAX_NAME 20
#define MAX_TITLE 40
#define MAX_TEXT 128
#define MAX_LIST 8
#define MAX_ITEM 24
#define MAX_Z 32
#define MAX_ICON 12
#define TITLE_H 38
#define TASK_H 48
#define TOP_H 0
#define FONT_SCALE 2
#define CW (8*FONT_SCALE)
#define CH (8*FONT_SCALE)
#define FB_MAX_W 1024
#define FB_MAX_H 768

#define CAP_PULSO_ENV 0x01u
#define CAP_PULSO_REC 0x02u
#define CAP_ORGES_CREAR 0x04u
#define CAP_DISP_LEER 0x80u
#define CAP_DISP_CTRL 0x200u
#define CAP_DESKTOP 0x400u
#define CAP_COMPOSE 0x800u
#define CAP_FB 0x1000u
#define CAP_ALL 0xFFFFFFFFu

#define K_MASTER 1
#define K_BUS 5
#define K_DISP 6
#define K_DESKTOP 7
#define K_FB 8
#define K_COMPOSE 9
#define K_WINDOW 10
#define K_TEXT 11
#define K_LIST 12
#define K_INSPECT 13
#define K_ICON 14

/* Paleta = desktop/shell/shell.css :root */
#define C_BG      0xFF070A10u
#define C_BG2     0xFF0C1220u
#define C_PANEL   0xFF121826u
#define C_TITLE   0xFF1A2233u
#define C_TITLE2  0xFF141A28u
#define C_BODY    0xFF0C1018u
#define C_LINE    0xFF2A3142u
#define C_LINE2   0xFF3A4258u
#define C_GOLD    0xFFF0C14Bu
#define C_ACCENT  0xFF5B9FD4u
#define C_TEXT    0xFFEEF2F8u
#define C_MUTED   0xFF8B93A7u
#define C_OK      0xFF3DDC97u
#define C_DANGER  0xFFE85D5Du
#define C_SHADOW  0x99000000u
#define C_CURSOR  0xFFFFFFFFu
#define C_TASK    0xF00E1420u
#define C_HOVER   0x14FFFFFFu

static inline u8 inb(u16 p){u8 v;__asm__ __volatile__("inb %1,%0":"=a"(v):"Nd"(p));return v;}
static inline void outb(u16 p,u8 v){__asm__ __volatile__("outb %0,%1"::"a"(v),"Nd"(p));}
static inline void outw(u16 p,u16 v){__asm__ __volatile__("outw %0,%1"::"a"(v),"Nd"(p));}
static inline u16 inw(u16 p){u16 v;__asm__ __volatile__("inw %1,%0":"=a"(v):"Nd"(p));return v;}
static void pausa(void){for(volatile int i=0;i<180;i++)__asm__ __volatile__("pause");}

struct OCB {
  u32 magic, oid, kind, caps, pin, pout;
  u8 alive, visible, focused, minimized;
  char nombre[MAX_NAME];
  char titulo[MAX_TITLE];
  int x, y, w, h, z;
  u32 color;
  char text[MAX_TEXT];
  int text_len;
  char items[MAX_LIST][MAX_ITEM];
  int n_items, sel_item;
  char glyph; /* icono escritorio */
};

static struct OCB ocbs[MAX_OCB];
static int n_ocb, n_pulso, focus_win=-1;
static int z_order[MAX_Z], n_z;
static u32 nodo_id, frame_tick;
static u32 *fb, fb_w, fb_h, fb_pitch;
static int fb_ok;
static u32 backbuf[FB_MAX_W*FB_MAX_H];
static u32 scene[FB_MAX_W*FB_MAX_H];
static int cur_ox=-1, cur_oy=-1, scene_valid=0;
static int mouse_x=100, mouse_y=100;
static u8 mouse_btn, mouse_cycle, mouse_pkt[3];
static int drag_win=-1, drag_ox, drag_oy, menu_open, orges_seq, phase;
/* phase 0=splash 1=desktop */
static int splash_ticks;
static volatile u16 *const VGA=(volatile u16*)0xB8000UL;

static const u8 FONT8[96][8]={
  {0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00},{0x18,0x3C,0x3C,0x18,0x18,0x00,0x18,0x00},
  {0x36,0x36,0x00,0x00,0x00,0x00,0x00,0x00},{0x36,0x36,0x7F,0x36,0x7F,0x36,0x36,0x00},
  {0x0C,0x3E,0x03,0x1E,0x30,0x1F,0x0C,0x00},{0x00,0x63,0x33,0x18,0x0C,0x66,0x63,0x00},
  {0x1C,0x36,0x1C,0x6E,0x3B,0x33,0x6E,0x00},{0x06,0x06,0x03,0x00,0x00,0x00,0x00,0x00},
  {0x18,0x0C,0x06,0x06,0x06,0x0C,0x18,0x00},{0x06,0x0C,0x18,0x18,0x18,0x0C,0x06,0x00},
  {0x00,0x66,0x3C,0xFF,0x3C,0x66,0x00,0x00},{0x00,0x0C,0x0C,0x3F,0x0C,0x0C,0x00,0x00},
  {0x00,0x00,0x00,0x00,0x00,0x0C,0x0C,0x06},{0x00,0x00,0x00,0x3F,0x00,0x00,0x00,0x00},
  {0x00,0x00,0x00,0x00,0x00,0x0C,0x0C,0x00},{0x60,0x30,0x18,0x0C,0x06,0x03,0x01,0x00},
  {0x3E,0x63,0x73,0x7B,0x6F,0x67,0x3E,0x00},{0x0C,0x0E,0x0C,0x0C,0x0C,0x0C,0x3F,0x00},
  {0x1E,0x33,0x30,0x1C,0x06,0x33,0x3F,0x00},{0x1E,0x33,0x30,0x1C,0x30,0x33,0x1E,0x00},
  {0x38,0x3C,0x36,0x33,0x7F,0x30,0x78,0x00},{0x3F,0x03,0x1F,0x30,0x30,0x33,0x1E,0x00},
  {0x1C,0x06,0x03,0x1F,0x33,0x33,0x1E,0x00},{0x3F,0x33,0x30,0x18,0x0C,0x0C,0x0C,0x00},
  {0x1E,0x33,0x33,0x1E,0x33,0x33,0x1E,0x00},{0x1E,0x33,0x33,0x3E,0x30,0x18,0x0E,0x00},
  {0x00,0x0C,0x0C,0x00,0x00,0x0C,0x0C,0x00},{0x00,0x0C,0x0C,0x00,0x00,0x0C,0x0C,0x06},
  {0x18,0x0C,0x06,0x03,0x06,0x0C,0x18,0x00},{0x00,0x00,0x3F,0x00,0x00,0x3F,0x00,0x00},
  {0x06,0x0C,0x18,0x30,0x18,0x0C,0x06,0x00},{0x1E,0x33,0x30,0x18,0x0C,0x00,0x0C,0x00},
  {0x3E,0x63,0x7B,0x7B,0x7B,0x03,0x1E,0x00},{0x0C,0x1E,0x33,0x33,0x3F,0x33,0x33,0x00},
  {0x3F,0x66,0x66,0x3E,0x66,0x66,0x3F,0x00},{0x3C,0x66,0x03,0x03,0x03,0x66,0x3C,0x00},
  {0x1F,0x36,0x66,0x66,0x66,0x36,0x1F,0x00},{0x7F,0x46,0x16,0x1E,0x16,0x46,0x7F,0x00},
  {0x7F,0x46,0x16,0x1E,0x16,0x06,0x0F,0x00},{0x3C,0x66,0x03,0x03,0x73,0x66,0x7C,0x00},
  {0x33,0x33,0x33,0x3F,0x33,0x33,0x33,0x00},{0x1E,0x0C,0x0C,0x0C,0x0C,0x0C,0x1E,0x00},
  {0x78,0x30,0x30,0x30,0x33,0x33,0x1E,0x00},{0x67,0x66,0x36,0x1E,0x36,0x66,0x67,0x00},
  {0x0F,0x06,0x06,0x06,0x46,0x66,0x7F,0x00},{0x63,0x77,0x7F,0x6B,0x63,0x63,0x63,0x00},
  {0x63,0x67,0x6F,0x7B,0x73,0x63,0x63,0x00},{0x1C,0x36,0x63,0x63,0x63,0x36,0x1C,0x00},
  {0x3F,0x66,0x66,0x3E,0x06,0x06,0x0F,0x00},{0x1E,0x33,0x33,0x33,0x3B,0x1E,0x38,0x00},
  {0x3F,0x66,0x66,0x3E,0x36,0x66,0x67,0x00},{0x1E,0x33,0x07,0x0E,0x38,0x33,0x1E,0x00},
  {0x3F,0x2D,0x0C,0x0C,0x0C,0x0C,0x1E,0x00},{0x33,0x33,0x33,0x33,0x33,0x33,0x3F,0x00},
  {0x33,0x33,0x33,0x33,0x33,0x1E,0x0C,0x00},{0x63,0x63,0x63,0x6B,0x7F,0x77,0x63,0x00},
  {0x63,0x63,0x36,0x1C,0x1C,0x36,0x63,0x00},{0x33,0x33,0x33,0x1E,0x0C,0x0C,0x1E,0x00},
  {0x7F,0x63,0x31,0x18,0x4C,0x66,0x7F,0x00},{0x1E,0x06,0x06,0x06,0x06,0x06,0x1E,0x00},
  {0x03,0x06,0x0C,0x18,0x30,0x60,0x40,0x00},{0x1E,0x18,0x18,0x18,0x18,0x18,0x1E,0x00},
  {0x08,0x1C,0x36,0x63,0x00,0x00,0x00,0x00},{0x00,0x00,0x00,0x00,0x00,0x00,0x00,0xFF},
  {0x0C,0x0C,0x18,0x00,0x00,0x00,0x00,0x00},{0x00,0x00,0x1E,0x30,0x3E,0x33,0x6E,0x00},
  {0x07,0x06,0x06,0x3E,0x66,0x66,0x3B,0x00},{0x00,0x00,0x1E,0x33,0x03,0x33,0x1E,0x00},
  {0x38,0x30,0x30,0x3e,0x33,0x33,0x6E,0x00},{0x00,0x00,0x1E,0x33,0x3f,0x03,0x1E,0x00},
  {0x1C,0x36,0x06,0x0f,0x06,0x06,0x0F,0x00},{0x00,0x00,0x6E,0x33,0x33,0x3E,0x30,0x1F},
  {0x07,0x06,0x36,0x6E,0x66,0x66,0x67,0x00},{0x0C,0x00,0x0E,0x0C,0x0C,0x0C,0x1E,0x00},
  {0x30,0x00,0x30,0x30,0x30,0x33,0x33,0x1E},{0x07,0x06,0x66,0x36,0x1E,0x36,0x67,0x00},
  {0x0E,0x0C,0x0C,0x0C,0x0C,0x0C,0x1E,0x00},{0x00,0x00,0x33,0x7F,0x7F,0x6B,0x63,0x00},
  {0x00,0x00,0x1F,0x33,0x33,0x33,0x33,0x00},{0x00,0x00,0x1E,0x33,0x33,0x33,0x1E,0x00},
  {0x00,0x00,0x3B,0x66,0x66,0x3E,0x06,0x0F},{0x00,0x00,0x6E,0x33,0x33,0x3E,0x30,0x78},
  {0x00,0x00,0x3B,0x6E,0x66,0x06,0x0F,0x00},{0x00,0x00,0x3E,0x03,0x1E,0x30,0x1F,0x00},
  {0x08,0x0C,0x3E,0x0C,0x0C,0x2C,0x18,0x00},{0x00,0x00,0x33,0x33,0x33,0x33,0x6E,0x00},
  {0x00,0x00,0x33,0x33,0x33,0x1E,0x0C,0x00},{0x00,0x00,0x63,0x6B,0x7F,0x7F,0x36,0x00},
  {0x00,0x00,0x63,0x36,0x1C,0x36,0x63,0x00},{0x00,0x00,0x33,0x33,0x33,0x3E,0x30,0x1F},
  {0x00,0x00,0x3F,0x19,0x0C,0x26,0x3F,0x00},{0x38,0x0C,0x0C,0x07,0x0C,0x0C,0x38,0x00},
  {0x18,0x18,0x18,0x00,0x18,0x18,0x18,0x00},{0x07,0x0C,0x0C,0x38,0x0C,0x0C,0x07,0x00},
  {0x6E,0x3B,0x00,0x00,0x00,0x00,0x00,0x00},{0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00}
};

static u32 fnv(const char *s){u32 h=2166136261u;while(*s){h^=(u8)*s++;h*=16777619u;}return h;}
static void cpy(char *d,const char *s,int n){int i=0;for(;i<n-1&&s[i];i++)d[i]=s[i];d[i]=0;}
static int igual(const char *a,const char *b){while(*a&&*b&&*a==*b){a++;b++;}return !*a&&!*b;}
static int buscar(const char *n){for(int i=0;i<n_ocb;i++)if(ocbs[i].alive&&igual(ocbs[i].nombre,n))return i;return -1;}
static int tiene_cap(int i,u32 c){if(i<0||!ocbs[i].alive)return 0;if(ocbs[i].kind==K_MASTER)return 1;return (ocbs[i].caps&c)!=0;}

#define VBE_IDX 0x01CE
#define VBE_DAT 0x01CF
static void vbe_write(u16 i,u16 v){outw(VBE_IDX,i);outw(VBE_DAT,v);}
static u16 vbe_read(u16 i){outw(VBE_IDX,i);return inw(VBE_DAT);}
static int fb_probe_addr(u32 addr,u32 pixels){
  volatile u32 *p=(volatile u32*)addr;p[0]=0xA11E70ADu;p[1]=0xB01DFACEu;
  if(p[0]!=0xA11E70ADu||p[1]!=0xB01DFACEu)return 0;
  for(u32 i=0;i<8&&i<pixels;i++)p[i]=C_GOLD;return 1;
}
static int fb_try_bases(u32 w,u32 h){
  static const u32 bases[]={0xE0000000u,0xFD000000u,0xF0000000u,0xD0000000u,0x80000000u,0};
  fb_w=w;fb_h=h;fb_pitch=w*4;if(w>FB_MAX_W||h>FB_MAX_H)return 0;
  for(int i=0;bases[i];i++)if(fb_probe_addr(bases[i],w*h)){fb=(u32*)bases[i];fb_ok=1;return 1;}
  return 0;
}
static int fb_init_bochs(int w,int h){
  vbe_write(0,0xB0C0);if(vbe_read(0)<0xB0C0)return 0;
  vbe_write(4,0);vbe_write(1,(u16)w);vbe_write(2,(u16)h);vbe_write(3,32);
  vbe_write(6,(u16)w);vbe_write(7,(u16)h);vbe_write(8,0);vbe_write(9,0);vbe_write(4,0x41);
  return fb_try_bases((u32)w,(u32)h);
}
static int fb_init_multiboot(u32 magic,u32 info){
  if(magic!=0x2BADB002u||!info)return 0;
  u32 *mi=(u32*)info;if(!(mi[0]&(1u<<12)))return 0;
  u32 *f=(u32*)(info+88);u32 addr=f[0];fb_pitch=f[2];fb_w=f[3];fb_h=f[4];
  u8 bpp=*((u8*)(info+108));
  if(!addr||fb_w<320||fb_w>FB_MAX_W||fb_h>FB_MAX_H||bpp!=32)return 0;
  if(!fb_probe_addr(addr,fb_w*fb_h))return 0;fb=(u32*)addr;fb_ok=1;return 1;
}
static int fb_init(u32 magic,u32 info){
  if(fb_init_multiboot(magic,info))return 1;
  if(fb_init_bochs(1024,768))return 1;
  if(fb_init_bochs(800,600))return 1;
  if(fb_init_bochs(640,480))return 1;
  return 0;
}

static u32 *bb(void){return backbuf;}
static void put_bb(int x,int y,u32 c){if(x<0||y<0||(u32)x>=fb_w||(u32)y>=fb_h)return;bb()[(u32)y*fb_w+(u32)x]=c;}
static void fill_bb(int x,int y,int w,int h,u32 c){
  if(w<=0||h<=0)return;
  if(x<0){w+=x;x=0;} if(y<0){h+=y;y=0;}
  if((u32)(x+w)>fb_w)w=(int)fb_w-x; if((u32)(y+h)>fb_h)h=(int)fb_h-y;
  if(w<=0||h<=0)return;
  for(int yy=0;yy<h;yy++){u32 *row=bb()+(u32)(y+yy)*fb_w+(u32)x;for(int xx=0;xx<w;xx++)row[xx]=c;}
}
static void blend_bb(int x,int y,int w,int h,u32 c){
  if(w<=0||h<=0)return;
  if(x<0){w+=x;x=0;} if(y<0){h+=y;y=0;}
  if((u32)(x+w)>fb_w)w=(int)fb_w-x; if((u32)(y+h)>fb_h)h=(int)fb_h-y;
  for(int yy=0;yy<h;yy++)for(int xx=0;xx<w;xx++){
    u32 *p=bb()+(u32)(y+yy)*fb_w+(u32)(x+xx);u32 d=*p;
    u32 r=(((d>>16)&0xFF)+((c>>16)&0xFF))>>1;
    u32 g=(((d>>8)&0xFF)+((c>>8)&0xFF))>>1;
    u32 b=((d&0xFF)+(c&0xFF))>>1;
    *p=0xFF000000u|(r<<16)|(g<<8)|b;
  }
}
static void rect_bb(int x,int y,int w,int h,u32 c){
  fill_bb(x,y,w,1,c);fill_bb(x,y+h-1,w,1,c);fill_bb(x,y,1,h,c);fill_bb(x+w-1,y,1,h,c);
}
/* panel con radio aproximado (shell --radius: 14px) */
static void panel_r(int x,int y,int w,int h,u32 fill,u32 border){
  fill_bb(x+3,y,w-6,h,fill);
  fill_bb(x,y+3,w,h-6,fill);
  fill_bb(x+1,y+1,w-2,h-2,fill);
  rect_bb(x,y,w,h,border);
  put_bb(x,y,C_BG);put_bb(x+w-1,y,C_BG);put_bb(x,y+h-1,C_BG);put_bb(x+w-1,y+h-1,C_BG);
  put_bb(x+1,y,border);put_bb(x,y+1,border);
  put_bb(x+w-2,y,border);put_bb(x+w-1,y+1,border);
}
static void draw_char(int x,int y,char ch,u32 fg){
  int idx=(int)(u8)ch-32;if(idx<0||idx>=96)idx=0;const u8 *g=FONT8[idx];
  for(int row=0;row<8;row++){u8 bits=g[row];
    for(int col=0;col<8;col++)if(bits&(1<<col))
      for(int sy=0;sy<FONT_SCALE;sy++)for(int sx=0;sx<FONT_SCALE;sx++)
        put_bb(x+col*FONT_SCALE+sx,y+row*FONT_SCALE+sy,fg);
  }
}
static void draw_text(int x,int y,const char *s,u32 fg){
  while(*s){if((u32)x+CW>fb_w)break;draw_char(x,y,*s++,fg);x+=CW;}
}
static void draw_text_clip(int x,int y,const char *s,u32 fg,int maxw){
  int cx=x;while(*s&&cx+CW<=x+maxw){draw_char(cx,y,*s++,fg);cx+=CW;}
}
static void present(void){
  if(!fb_ok)return;
  for(u32 y=0;y<fb_h;y++){
    u32 *src=bb()+y*fb_w;u32 *dst=(u32*)((u8*)fb+y*fb_pitch);
    for(u32 x=0;x<fb_w;x++)dst[x]=src[x];
  }
}

static int ocb_add(const char *n,u32 kind,u32 caps){
  if(n_ocb>=MAX_OCB)return -1;if(buscar(n)>=0)return -2;
  struct OCB *o=&ocbs[n_ocb];
  o->magic=0x4F434200;o->oid=fnv(n)^nodo_id^(u32)n_ocb;o->kind=kind;o->caps=caps;
  o->pin=o->pout=0;o->alive=1;o->x=80;o->y=48;o->w=360;o->h=240;o->z=0;
  o->visible=1;o->focused=0;o->minimized=0;o->color=C_PANEL;
  o->text[0]=0;o->text_len=0;o->n_items=0;o->sel_item=0;o->titulo[0]=0;o->glyph='?';
  cpy(o->nombre,n,MAX_NAME);return n_ocb++;
}
static int pulso(int from,int to){
  if(from<0||to<0)return -1;if(!tiene_cap(from,CAP_PULSO_ENV))return -2;
  if(!tiene_cap(to,CAP_PULSO_REC))return -3;
  ocbs[from].pout++;ocbs[to].pin++;n_pulso++;return 0;
}
static int is_win(int i){
  if(i<0||i>=n_ocb||!ocbs[i].alive)return 0;
  u32 k=ocbs[i].kind;return k==K_WINDOW||k==K_TEXT||k==K_LIST||k==K_INSPECT;
}
static void z_raise(int i){
  if(!is_win(i)||ocbs[i].minimized)return;
  int j,k=0,tmp[MAX_Z];
  for(j=0;j<n_z;j++)if(z_order[j]!=i)tmp[k++]=z_order[j];
  tmp[k++]=i;n_z=k;for(j=0;j<n_z;j++){z_order[j]=tmp[j];ocbs[tmp[j]].z=j;}
}
static void z_add(int i){if(n_z>=MAX_Z)return;z_order[n_z++]=i;ocbs[i].z=n_z-1;}
static void focus_set(int i){
  for(int j=0;j<n_ocb;j++)ocbs[j].focused=0;
  if(i>=0&&is_win(i)&&!ocbs[i].minimized){ocbs[i].focused=1;focus_win=i;z_raise(i);}
  else focus_win=-1;
}
static void text_append(int i,char ch){
  if(i<0||ocbs[i].kind!=K_TEXT)return;if(ocbs[i].text_len>=MAX_TEXT-1)return;
  ocbs[i].text[ocbs[i].text_len++]=ch;ocbs[i].text[ocbs[i].text_len]=0;
}
static void text_backspace(int i){
  if(i<0||ocbs[i].kind!=K_TEXT||ocbs[i].text_len<=0)return;ocbs[i].text[--ocbs[i].text_len]=0;
}
static int hit_win(int mx,int my){
  for(int zi=n_z-1;zi>=0;zi--){
    int i=z_order[zi];
    if(!ocbs[i].alive||!ocbs[i].visible||ocbs[i].minimized)continue;
    if(mx>=ocbs[i].x&&mx<ocbs[i].x+ocbs[i].w&&my>=ocbs[i].y&&my<ocbs[i].y+ocbs[i].h)return i;
  }
  return -1;
}
static int hit_icon(int mx,int my){
  for(int i=0;i<n_ocb;i++){
    if(!ocbs[i].alive||ocbs[i].kind!=K_ICON)continue;
    if(mx>=ocbs[i].x&&mx<ocbs[i].x+88&&my>=ocbs[i].y&&my<ocbs[i].y+90)return i;
  }
  return -1;
}

static int crear_orges(u32 kind,const char *base){
  char name[MAX_NAME];int n=0;while(base[n]&&n<10){name[n]=base[n];n++;}
  name[n++]='0'+(orges_seq%10);name[n]=0;orges_seq++;
  int i=ocb_add(name,kind,CAP_PULSO_ENV|CAP_PULSO_REC|CAP_DESKTOP);if(i<0)return i;
  if(kind==K_TEXT){cpy(ocbs[i].titulo,"Editor · campo",MAX_TITLE);ocbs[i].w=400;ocbs[i].h=260;}
  else if(kind==K_LIST){
    cpy(ocbs[i].titulo,"Organismos",MAX_TITLE);ocbs[i].n_items=4;
    cpy(ocbs[i].items[0],"Maestro",MAX_ITEM);cpy(ocbs[i].items[1],"Compositor",MAX_ITEM);
    cpy(ocbs[i].items[2],"Desktop",MAX_ITEM);cpy(ocbs[i].items[3],"Framebuffer",MAX_ITEM);
    ocbs[i].w=320;ocbs[i].h=280;
  }else if(kind==K_INSPECT){cpy(ocbs[i].titulo,"Estado · red OCB",MAX_TITLE);ocbs[i].w=380;ocbs[i].h=300;}
  else cpy(ocbs[i].titulo,"Ventana Alset",MAX_TITLE);
  ocbs[i].x=70+(orges_seq%4)*28;ocbs[i].y=40+(orges_seq%3)*24;
  z_add(i);focus_set(i);return i;
}
static void launch_from_icon(int ii){
  char g=ocbs[ii].glyph;
  if(g=='S'||g=='s') crear_orges(K_WINDOW,"Studio");
  else if(g=='E'||g=='e') crear_orges(K_TEXT,"Editor");
  else if(g=='F'||g=='f') crear_orges(K_LIST,"Files");
  else if(g=='T'||g=='t') crear_orges(K_TEXT,"Term");
  else if(g=='M'||g=='m') crear_orges(K_INSPECT,"Mind");
  else if(g=='O'||g=='o') crear_orges(K_INSPECT,"Orges");
  else crear_orges(K_WINDOW,"App");
}

/* —— Dibujo estilo shell.css —— */
static void draw_wallpaper(void){
  /* radial + linear como #os en shell.css */
  for(u32 y=0;y<fb_h;y++){
    u32 t=(y*255)/(fb_h?fb_h:1);
    u32 r=7+(t*5)/255,g=10+(t*8)/255,b=16+(t*12)/255;
    u32 c=0xFF000000u|(r<<16)|(g<<8)|b;
    u32 *row=bb()+y*fb_w;
    for(u32 x=0;x<fb_w;x++)row[x]=c;
  }
  /* elipse azul (accent) esquina superior */
  int cx=(int)fb_w/5,cy=(int)fb_h/8,rad=(int)fb_w/3;
  for(int dy=-rad/2;dy<=rad/2;dy++)for(int dx=-rad;dx<=rad;dx++){
    int d2=dx*dx+dy*dy*4;if(d2>rad*rad)continue;
    u32 a=(u32)(rad*rad-d2)*18/(u32)(rad*rad+1);
    int px=cx+dx,py=cy+dy;if(px<0||py<0||(u32)px>=fb_w||(u32)py>=fb_h)continue;
    u32 *p=bb()+(u32)py*fb_w+(u32)px;u32 d=*p;
    u32 r=((d>>16)&0xFF);u32 g=((d>>8)&0xFF)+a/2;u32 b=(d&0xFF)+a;
    if(g>40)g=40;if(b>55)b=55;*p=0xFF000000u|(r<<16)|(g<<8)|b;
  }
  /* elipse oro inferior derecha */
  cx=(int)fb_w*9/10;cy=(int)fb_h*4/5;rad=(int)fb_w/5;
  for(int dy=-rad/2;dy<=rad/2;dy++)for(int dx=-rad;dx<=rad;dx++){
    int d2=dx*dx+dy*dy*3;if(d2>rad*rad)continue;
    u32 a=(u32)(rad*rad-d2)*14/(u32)(rad*rad+1);
    int px=cx+dx,py=cy+dy;if(px<0||py<0||(u32)px>=fb_w||(u32)py>=fb_h)continue;
    u32 *p=bb()+(u32)py*fb_w+(u32)px;u32 d=*p;
    u32 r=((d>>16)&0xFF)+a;u32 g=((d>>8)&0xFF)+a/2;u32 b=(d&0xFF);
    if(r>50)r=50;if(g>40)g=40;*p=0xFF000000u|(r<<16)|(g<<8)|b;
  }
}
static void draw_icon(int i){
  struct OCB *o=&ocbs[i];
  int x=o->x,y=o->y;
  /* glyph box 52x52 */
  panel_r(x+18,y,52,52,0xFF1C2436u,C_LINE2);
  char gl[2]={o->glyph,0};
  draw_text(x+18+18,y+18,gl,C_GOLD);
  draw_text_clip(x,y+58,o->titulo[0]?o->titulo:o->nombre,C_TEXT,88);
}
static void draw_window(int i){
  struct OCB *o=&ocbs[i];
  int x=o->x,y=o->y,w=o->w,h=o->h;
  blend_bb(x+8,y+12,w,h,C_SHADOW);
  u32 bd=o->focused?0xFF5A4A28u:C_LINE2; /* oro suave si foco */
  panel_r(x,y,w,h,C_PANEL,bd);
  /* titlebar gradient approx */
  fill_bb(x+1,y+1,w-2,TITLE_H,C_TITLE);
  fill_bb(x+1,y+TITLE_H/2,w-2,TITLE_H/2,C_TITLE2);
  fill_bb(x+1,y+TITLE_H,w-2,1,C_LINE);
  if(o->focused) fill_bb(x+1,y+TITLE_H-1,w-2,1,0xFF6A5A30u);
  draw_text_clip(x+12,y+11,o->titulo[0]?o->titulo:o->nombre,C_TEXT,w-100);
  /* botones estilo shell: close danger hover */
  panel_r(x+w-34,y+7,26,22,0xFF1A2030u,C_LINE);
  fill_bb(x+w-28,y+15,14,2,C_DANGER);
  panel_r(x+w-64,y+7,26,22,0xFF1A2030u,C_LINE);
  fill_bb(x+w-56,y+16,10,2,C_MUTED);
  panel_r(x+w-94,y+7,26,22,0xFF1A2030u,C_LINE);
  fill_bb(x+w-86,y+12,10,10,C_OK);
  /* body */
  fill_bb(x+1,y+TITLE_H+1,w-2,h-TITLE_H-2,C_BODY);
  int cx=x+14,cy=y+TITLE_H+12,cw=w-28;
  if(o->kind==K_TEXT){
    draw_text(cx,cy,"Campo de texto · organismo",C_MUTED);
    panel_r(cx,cy+CH+8,cw,h-TITLE_H-CH-36,0xFF080B12u,C_LINE);
    draw_text_clip(cx+10,cy+CH+16,o->text,C_TEXT,cw-20);
    if(o->focused&&((frame_tick/18)&1))
      fill_bb(cx+10+o->text_len*CW,cy+CH+16,2,CH,C_GOLD);
  }else if(o->kind==K_LIST){
    draw_text(cx,cy,"Red de organismos",C_MUTED);
    for(int k=0;k<o->n_items;k++){
      int iy=cy+CH+10+k*(CH+10);
      if(iy+CH>y+h-12)break;
      u32 bg=(k==o->sel_item)?0xFF1A2840u:0xFF101620u;
      panel_r(cx,iy-4,cw,CH+8,bg,k==o->sel_item?C_GOLD:C_LINE);
      draw_text_clip(cx+12,iy,o->items[k],C_TEXT,cw-24);
    }
  }else if(o->kind==K_INSPECT){
    draw_text(cx,cy,"Inspector · OCB vivos",C_MUTED);
    int row=0;
    for(int j=0;j<n_ocb&&row<9;j++){
      if(!ocbs[j].alive)continue;
      char line[36];int n=0;const char *nm=ocbs[j].nombre;
      while(*nm&&n<14)line[n++]=*nm++;
      line[n++]=' ';line[n++]='#';
      const char *hex="0123456789ABCDEF";u32 id=ocbs[j].oid&0xFF;
      line[n++]=hex[(id>>4)&0xF];line[n++]=hex[id&0xF];line[n]=0;
      draw_text(cx,cy+CH+8+row*(CH+4),line,ocbs[j].focused?C_GOLD:C_TEXT);
      row++;
    }
  }else{
    draw_text(cx,cy,"Alset Desktop · orges",C_MUTED);
    draw_text(cx,cy+CH+10,"Ventana como organismo.",C_TEXT);
    draw_text(cx,cy+2*(CH+10),"Caps · pulsos · RootCID local.",C_MUTED);
    panel_r(cx,cy+3*(CH+12),120,48,0xFF1C2436u,C_GOLD);
    draw_text(cx+20,cy+3*(CH+12)+16,"ORGES",C_GOLD);
  }
}
static void draw_taskbar(void){
  int y=(int)fb_h-TASK_H;
  fill_bb(0,y,(int)fb_w,TASK_H,C_TASK);
  fill_bb(0,y,(int)fb_w,1,C_LINE);
  /* start button oro */
  panel_r(8,y+8,40,32,menu_open?0xFF2A2410u:0xFF1A2030u,menu_open?C_GOLD:C_LINE2);
  draw_text(16,y+14,"A",C_GOLD);
  /* task buttons for windows */
  int tx=56;
  for(int zi=0;zi<n_z;zi++){
    int i=z_order[zi];
    if(!ocbs[i].alive||!is_win(i))continue;
    int tw=100;
    if(tx+tw>(int)fb_w-120)break;
    u32 bg=ocbs[i].focused?0xFF1A2840u:0xFF121820u;
    u32 bd=ocbs[i].focused?C_GOLD:C_LINE;
    panel_r(tx,y+8,tw,32,bg,bd);
    draw_text_clip(tx+8,y+14,ocbs[i].titulo[0]?ocbs[i].titulo:ocbs[i].nombre,C_TEXT,tw-16);
    tx+=tw+6;
  }
  /* clock / status derecha */
  char buf[24];int n=0;
  buf[n++]='o';buf[n++]=':';
  int alive=0;for(int i=0;i<n_ocb;i++)if(ocbs[i].alive)alive++;
  if(alive>=10){buf[n++]=(char)('0'+alive/10);buf[n++]=(char)('0'+alive%10);}else buf[n++]=(char)('0'+alive);
  buf[n++]=' ';buf[n++]='p';buf[n++]=':';
  if(n_pulso>=10){buf[n++]=(char)('0'+(n_pulso/10)%10);buf[n++]=(char)('0'+n_pulso%10);}else buf[n++]=(char)('0'+n_pulso);
  buf[n]=0;
  panel_r((int)fb_w-12-n*CW-16,y+8,n*CW+16,32,0xFF121820u,C_LINE);
  draw_text((int)fb_w-12-n*CW-8,y+14,buf,C_MUTED);
}
static void draw_start_menu(void){
  if(!menu_open)return;
  int mw=240,mh=280,mx=8,my=(int)fb_h-TASK_H-mh-8;
  blend_bb(mx+6,my+8,mw,mh,C_SHADOW);
  panel_r(mx,my,mw,mh,C_PANEL,C_LINE2);
  fill_bb(mx+1,my+1,mw-2,36,C_TITLE);
  draw_text(mx+14,my+10,"ALSET OS",C_GOLD);
  draw_text(mx+14+8*CW,my+10,"Desktop",C_MUTED);
  const char *items[]={"Studio","Editor","Archivos","Terminal","Mind","Organismo","Estado",0};
  for(int i=0;items[i];i++){
    int iy=my+44+i*(CH+12);
    draw_text(mx+16,iy,items[i],C_TEXT);
  }
  fill_bb(mx+12,my+mh-48,mw-24,1,C_LINE);
  draw_text(mx+16,my+mh-36,"1-4 crear · M menu",C_MUTED);
}
static void draw_splash(void){
  fill_bb(0,0,(int)fb_w,(int)fb_h,C_BG);
  /* centro */
  int cx=(int)fb_w/2,cy=(int)fb_h/2;
  /* marca oro */
  panel_r(cx-36,cy-90,72,72,0xFF1A2030u,C_GOLD);
  draw_text(cx-8,cy-66,"A",C_GOLD);
  draw_text(cx-7*CW,cy, "ALSETOS",C_TEXT);
  draw_text(cx-11*CW,cy+CH+12,"Desktop · organismos",C_MUTED);
  /* barra carga */
  int bw=200,bx=cx-bw/2,by=cy+CH*3;
  fill_bb(bx,by,bw,3,0xFF1A2030u);
  int prog=(splash_ticks%40)*bw/40;
  fill_bb(bx,by,prog,3,C_GOLD);
}

static void scene_store(void){
  u32 n=fb_w*fb_h;for(u32 i=0;i<n;i++)scene[i]=backbuf[i];scene_valid=1;
}
static void scene_restore_rect(int x,int y,int w,int h){
  if(!scene_valid||!fb_ok)return;
  if(x<0){w+=x;x=0;}if(y<0){h+=y;y=0;}
  if((u32)(x+w)>fb_w)w=(int)fb_w-x;if((u32)(y+h)>fb_h)h=(int)fb_h-y;
  for(int yy=0;yy<h;yy++){
    u32 *s=scene+(u32)(y+yy)*fb_w+(u32)x;
    u32 *d=(u32*)((u8*)fb+(u32)(y+yy)*fb_pitch)+x;
    for(int xx=0;xx<w;xx++)d[xx]=s[xx];
  }
}
static void cursor_draw_on_fb(int x,int y){
  for(int i=0;i<14;i++){
    if(x>=0&&y+i>=0&&(u32)x<fb_w&&(u32)(y+i)<fb_h)
      ((u32*)((u8*)fb+(u32)(y+i)*fb_pitch))[x]=C_CURSOR;
  }
  for(int i=0;i<10;i++){
    int px=x+i,py=y+i;
    if(px>=0&&py>=0&&(u32)px<fb_w&&(u32)py<fb_h)
      ((u32*)((u8*)fb+(u32)py*fb_pitch))[px]=C_CURSOR;
  }
  for(int i=1;i<7;i++){
    int px=x+1,py=y+i;
    if(px>=0&&py>=0&&(u32)px<fb_w&&(u32)py<fb_h)
      ((u32*)((u8*)fb+(u32)py*fb_pitch))[px]=C_GOLD;
  }
}
static void cursor_move_fast(int nx,int ny){
  if(scene_valid&&cur_ox>=0)scene_restore_rect(cur_ox-2,cur_oy-2,20,20);
  cursor_draw_on_fb(nx,ny);cur_ox=nx;cur_oy=ny;
}
static void compositor_tick(void){
  if(!fb_ok)return;frame_tick++;
  if(phase==0){draw_splash();present();return;}
  draw_wallpaper();
  for(int i=0;i<n_ocb;i++)if(ocbs[i].alive&&ocbs[i].kind==K_ICON)draw_icon(i);
  for(int zi=0;zi<n_z;zi++){
    int i=z_order[zi];
    if(ocbs[i].alive&&ocbs[i].visible&&!ocbs[i].minimized)draw_window(i);
  }
  draw_taskbar();draw_start_menu();
  scene_store();present();
  cursor_draw_on_fb(mouse_x,mouse_y);cur_ox=mouse_x;cur_oy=mouse_y;
}

static void mouse_wait(u8 type){
  for(int i=0;i<100000;i++){
    if(type==1&&(inb(0x64)&1))return;
    if(type==2&&!(inb(0x64)&2))return;
  }
}
static void mouse_write(u8 val){mouse_wait(2);outb(0x64,0xD4);mouse_wait(2);outb(0x60,val);}
static u8 mouse_read(void){mouse_wait(1);return inb(0x60);}
static void mouse_init(void){
  mouse_wait(2);outb(0x64,0xA8);mouse_wait(2);outb(0x64,0x20);mouse_wait(1);
  u8 status=inb(0x60);status|=0x02;status&=(u8)~0x20;
  mouse_wait(2);outb(0x64,0x60);mouse_wait(2);outb(0x60,status);
  mouse_write(0xF6);mouse_read();mouse_write(0xF4);mouse_read();
  mouse_cycle=0;mouse_x=(int)fb_w/2;mouse_y=(int)fb_h/2;
}
static void mouse_feed(u8 sc){
  if(mouse_cycle==0&&!(sc&0x08))return;
  mouse_pkt[mouse_cycle++]=sc;if(mouse_cycle<3)return;mouse_cycle=0;
  u8 b0=mouse_pkt[0];
  int dx=(int)(i8)mouse_pkt[1];int dy=(int)(i8)mouse_pkt[2];
  if(dx>1||dx<-1)dx*=2;if(dy>1||dy<-1)dy*=2;
  mouse_x+=dx;mouse_y-=dy;
  if(mouse_x<0)mouse_x=0;if(mouse_y<0)mouse_y=0;
  if((u32)mouse_x>=fb_w)mouse_x=(int)fb_w-1;
  if((u32)mouse_y>=fb_h)mouse_y=(int)fb_h-1;
  u8 prev=mouse_btn;mouse_btn=b0&0x07;
  int need_scene=0;
  if(phase==0){if((mouse_btn&1)&&!(prev&1)){phase=1;need_scene=1;}}
  else if((mouse_btn&1)&&!(prev&1)){
    int ty=(int)fb_h-TASK_H;
    if(mouse_y>=ty){
      if(mouse_x>=8&&mouse_x<48){menu_open=!menu_open;need_scene=1;}
      else{
        int tx=56;
        for(int zi=0;zi<n_z;zi++){
          int i=z_order[zi];if(!ocbs[i].alive||!is_win(i))continue;
          if(mouse_x>=tx&&mouse_x<tx+100){
            if(ocbs[i].minimized)ocbs[i].minimized=0;
            focus_set(i);need_scene=1;break;
          }
          tx+=106;
        }
      }
    }else if(menu_open){
      int mw=240,mh=280,mx=8,my=ty-mh-8;
      if(mouse_x>=mx&&mouse_x<mx+mw&&mouse_y>=my&&mouse_y<my+mh){
        int idx=(mouse_y-(my+44))/(CH+12);
        if(idx==0)crear_orges(K_WINDOW,"Studio");
        else if(idx==1)crear_orges(K_TEXT,"Editor");
        else if(idx==2)crear_orges(K_LIST,"Files");
        else if(idx==3)crear_orges(K_TEXT,"Term");
        else if(idx==4||idx==5)crear_orges(K_INSPECT,"Mind");
        else if(idx==6)crear_orges(K_INSPECT,"Stat");
        menu_open=0;need_scene=1;
      }else{menu_open=0;need_scene=1;}
    }else{
      int hi=hit_win(mouse_x,mouse_y);
      if(hi>=0){
        focus_set(hi);
        if(mouse_y<ocbs[hi].y+TITLE_H){
          if(mouse_x>=ocbs[hi].x+ocbs[hi].w-34){
            ocbs[hi].alive=0;ocbs[hi].visible=0;
            int k=0;for(int j=0;j<n_z;j++)if(z_order[j]!=hi)z_order[k++]=z_order[j];
            n_z=k;focus_win=-1;
          }else if(mouse_x>=ocbs[hi].x+ocbs[hi].w-64){
            ocbs[hi].minimized=1;focus_win=-1;
          }else{drag_win=hi;drag_ox=mouse_x-ocbs[hi].x;drag_oy=mouse_y-ocbs[hi].y;}
        }
        need_scene=1;
      }else{
        int ic=hit_icon(mouse_x,mouse_y);
        if(ic>=0){launch_from_icon(ic);need_scene=1;}
      }
    }
  }
  if(!(mouse_btn&1))drag_win=-1;
  if(drag_win>=0&&ocbs[drag_win].alive){
    ocbs[drag_win].x=mouse_x-drag_ox;ocbs[drag_win].y=mouse_y-drag_oy;
    if(ocbs[drag_win].y<0)ocbs[drag_win].y=0;
    if(ocbs[drag_win].y>(int)fb_h-TASK_H-40)ocbs[drag_win].y=(int)fb_h-TASK_H-40;
    if(ocbs[drag_win].x<0)ocbs[drag_win].x=0;
    if(ocbs[drag_win].x>(int)fb_w-40)ocbs[drag_win].x=(int)fb_w-40;
    need_scene=1;
  }
  if(need_scene)compositor_tick();
  else if(phase==1)cursor_move_fast(mouse_x,mouse_y);
}

static char sc_ascii(u8 sc){
  static const char map[0x40]={
    0,0,'1','2','3','4','5','6','7','8','9','0','-','=',0,0,
    'q','w','e','r','t','y','u','i','o','p','[',']',0,0,
    'a','s','d','f','g','h','j','k','l',';','\'','`',0,'\\',
    'z','x','c','v','b','n','m',',','.','/',0,0,0,' '
  };
  if(sc<0x40)return map[sc];return 0;
}
static void vga_msg(const char *s){int i=0;while(s[i]&&i<80){VGA[i]=(u16)s[i]|0x0A00;i++;}}

static void add_icon(const char *name,const char *title,char glyph,int x,int y){
  int i=ocb_add(name,K_ICON,CAP_DESKTOP);if(i<0)return;
  cpy(ocbs[i].titulo,title,MAX_TITLE);ocbs[i].glyph=glyph;ocbs[i].x=x;ocbs[i].y=y;
  ocbs[i].w=88;ocbs[i].h=90;
}
static void bootstrap_organisms(void){
  nodo_id=0xA15E7001u;
  ocb_add("Maestro",K_MASTER,CAP_ALL);
  ocb_add("Framebuffer",K_FB,CAP_FB|CAP_DISP_LEER);
  ocb_add("Compositor",K_COMPOSE,CAP_COMPOSE|CAP_PULSO_ENV|CAP_PULSO_REC);
  ocb_add("Desktop",K_DESKTOP,CAP_DESKTOP|CAP_PULSO_ENV|CAP_PULSO_REC|CAP_ORGES_CREAR);
  ocb_add("Bus",K_BUS,CAP_PULSO_ENV|CAP_PULSO_REC);
  ocb_add("Teclado",K_DISP,CAP_DISP_LEER|CAP_DISP_CTRL);
  ocb_add("Raton",K_DISP,CAP_DISP_LEER|CAP_DISP_CTRL);
  /* iconos como en shell.js DEFAULT_ICONS */
  add_icon("icStudio","Studio",'S',24,24);
  add_icon("icEditor","Editor",'E',24,120);
  add_icon("icFiles","Archivos",'F',24,216);
  add_icon("icTerm","Terminal",'T',24,312);
  add_icon("icMind","Mind",'M',120,24);
  add_icon("icOrges","Organismo",'O',120,120);
}

void kernel_main(u32 magic,u32 info){
  (void)info;
  for(int i=0;i<80*25;i++)VGA[i]=0x0F00|' ';
  vga_msg("AlsetOS Desktop nativo...");
  if(!fb_init(magic,info)){vga_msg("Sin framebuffer");for(;;)pausa();}
  mouse_x=(int)fb_w/2;mouse_y=(int)fb_h/2;
  bootstrap_organisms();
  mouse_init();
  phase=0;splash_ticks=0;
  compositor_tick();
  for(;;){
    if(!(inb(0x64)&1)){
      if(phase==0){
        splash_ticks++;
        if(splash_ticks>90){phase=1;compositor_tick();}
        else if((splash_ticks%4)==0)compositor_tick();
        else pausa();
      }else if(focus_win>=0&&ocbs[focus_win].kind==K_TEXT&&(frame_tick%30)==0)
        compositor_tick();
      else pausa();
      frame_tick++;continue;
    }
    u8 st=inb(0x64);u8 sc=inb(0x60);
    if(st&0x20){mouse_feed(sc);continue;}
    if(sc==0xE0||(sc&0x80))continue;
    if(phase==0){phase=1;compositor_tick();continue;}
    if(sc==0x32){menu_open=!menu_open;compositor_tick();continue;}
    if(sc==0x02){crear_orges(K_WINDOW,"Win");menu_open=0;compositor_tick();continue;}
    if(sc==0x03){crear_orges(K_TEXT,"Txt");menu_open=0;compositor_tick();continue;}
    if(sc==0x04){crear_orges(K_LIST,"Lst");menu_open=0;compositor_tick();continue;}
    if(sc==0x05){crear_orges(K_INSPECT,"Ins");menu_open=0;compositor_tick();continue;}
    if(sc==0x1C){int from=buscar("Desktop");if(from>=0&&focus_win>=0)pulso(from,focus_win);compositor_tick();continue;}
    if(sc==0x0F){
      int start=focus_win+1;
      for(int k=0;k<n_ocb;k++){int i=(start+k)%n_ocb;if(is_win(i)&&!ocbs[i].minimized){focus_set(i);break;}}
      compositor_tick();continue;
    }
    if(sc==0x0E){text_backspace(focus_win);compositor_tick();continue;}
    if(focus_win>=0&&ocbs[focus_win].kind==K_LIST){
      if(sc==0x48&&ocbs[focus_win].sel_item>0){ocbs[focus_win].sel_item--;compositor_tick();continue;}
      if(sc==0x50&&ocbs[focus_win].sel_item<ocbs[focus_win].n_items-1){ocbs[focus_win].sel_item++;compositor_tick();continue;}
    }
    char ch=sc_ascii(sc);
    if(ch&&focus_win>=0&&ocbs[focus_win].kind==K_TEXT){text_append(focus_win,ch);compositor_tick();}
  }
}
