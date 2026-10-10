/*
 * AlsetOS Kernel — Desktop nativo por organismos
 * Maestro → Framebuffer → Compositor → Desktop → Orges
 * Doble buffer, fuente 2x, raton pantalla completa, z-order, dock, sombras.
 */
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long u64;
typedef signed char i8;

#define MAX_OCB 64
#define MAX_NAME 20
#define MAX_TITLE 36
#define MAX_TEXT 128
#define MAX_LIST 8
#define MAX_ITEM 24
#define MAX_Z 32
#define TITLE_H 28
#define DOCK_H 52
#define TOP_H 32
#define FONT_SCALE 2
#define CW (8*FONT_SCALE)
#define CH (8*FONT_SCALE)
#define FB_MAX_W 1024
#define FB_MAX_H 768

#define CAP_PULSO_ENV 0x01u
#define CAP_PULSO_REC 0x02u
#define CAP_ORGES_CREAR 0x04u
#define CAP_DISP_LEER 0x80u
#define CAP_DISP_ESCR 0x100u
#define CAP_DISP_CTRL 0x200u
#define CAP_DESKTOP 0x400u
#define CAP_COMPOSE 0x800u
#define CAP_FB 0x1000u
#define CAP_ALL 0xFFFFFFFFu

#define K_MASTER 1
#define K_SHELL 2
#define K_ORGES 3
#define K_BUS 5
#define K_DISP 6
#define K_DESKTOP 7
#define K_FB 8
#define K_COMPOSE 9
#define K_WINDOW 10
#define K_TEXT 11
#define K_LIST 12
#define K_INSPECT 13

#define C_BG0 0xFF070B14u
#define C_TOP 0xE0142238u
#define C_DOCK 0xF0121E32u
#define C_ACCENT 0xFF3DDC97u
#define C_ACCENT2 0xFF5B8DEEu
#define C_TEXT 0xFFF0F4FAu
#define C_DIM 0xFF8A9BB5u
#define C_MUTED 0xFF5A6A82u
#define C_WIN 0xFF182538u
#define C_TITLE 0xFF1E3048u
#define C_TITLE_F 0xFF243A58u
#define C_DANGER 0xFFE05C5Cu
#define C_WARN 0xFFF0B429u
#define C_CURSOR 0xFFFFFFFFu
#define C_SHADOW 0x88000000u
#define C_SEL 0xFF2A4A6Eu
#define C_BORDER 0xFF3A5080u
#define C_ICON_BG 0xFF1A3050u
#define C_PANEL 0xFF152238u
#define C_PANEL2 0xFF1B2C48u

static inline u8 inb(u16 p){u8 v;__asm__ __volatile__("inb %1,%0":"=a"(v):"Nd"(p));return v;}
static inline void outb(u16 p,u8 v){__asm__ __volatile__("outb %0,%1"::"a"(v),"Nd"(p));}
static inline void outw(u16 p,u16 v){__asm__ __volatile__("outw %0,%1"::"a"(v),"Nd"(p));}
static inline u16 inw(u16 p){u16 v;__asm__ __volatile__("inw %1,%0":"=a"(v):"Nd"(p));return v;}
static void pausa(void){for(volatile int i=0;i<200;i++)__asm__ __volatile__("pause");}

struct OCB {
    u32 magic, oid, kind, caps, pin, pout;
    u8 alive, visible, focused, minimized;
    char nombre[MAX_NAME];
    char titulo[MAX_TITLE];
    int x, y, w, h, z;
    u32 color, disp_clase, disp_dato;
    char text[MAX_TEXT];
    int text_len;
    char items[MAX_LIST][MAX_ITEM];
    int n_items, sel_item;
};

static struct OCB ocbs[MAX_OCB];
static int n_ocb, n_pulso, focus_win = -1;
static int z_order[MAX_Z];
static int n_z;
static u32 nodo_id, nonce, frame_tick;
static u32 *fb;
static u32 fb_w, fb_h, fb_pitch;
static int fb_ok;
static u32 backbuf[FB_MAX_W * FB_MAX_H];
static int mouse_x = 100, mouse_y = 100;
static u8 mouse_btn, mouse_cycle, mouse_pkt[3];
static int drag_win = -1, drag_ox, drag_oy, menu_open, orges_seq;
static volatile u16 *const VGA = (volatile u16 *)0xB8000UL;

static const u8 FONT8[96][8] = {
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
  volatile u32 *p=(volatile u32*)addr;
  p[0]=0xA11E70ADu;p[1]=0xB01DFACEu;
  if(p[0]!=0xA11E70ADu||p[1]!=0xB01DFACEu)return 0;
  for(u32 i=0;i<16&&i<pixels;i++)p[i]=0xFF3DDC97u;
  return 1;
}
static int fb_try_bases(u32 w,u32 h){
  static const u32 bases[]={0xE0000000u,0xFD000000u,0xF0000000u,0xD0000000u,0x80000000u,0};
  fb_w=w;fb_h=h;fb_pitch=w*4;
  if(w>FB_MAX_W||h>FB_MAX_H)return 0;
  u32 pixels=w*h;
  for(int i=0;bases[i];i++)if(fb_probe_addr(bases[i],pixels)){fb=(u32*)bases[i];fb_ok=1;return 1;}
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
  u32 *f=(u32*)(info+88);u32 addr=f[0];
  fb_pitch=f[2];fb_w=f[3];fb_h=f[4];
  u8 bpp=*((u8*)(info+108));
  if(!addr||fb_w<320||fb_w>FB_MAX_W||fb_h>FB_MAX_H||bpp!=32)return 0;
  if(!fb_probe_addr(addr,fb_w*fb_h))return 0;
  fb=(u32*)addr;fb_ok=1;return 1;
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
  if(x<0){w+=x;x=0;}if(y<0){h+=y;y=0;}
  if((u32)(x+w)>fb_w)w=(int)fb_w-x;if((u32)(y+h)>fb_h)h=(int)fb_h-y;
  if(w<=0||h<=0)return;
  for(int yy=0;yy<h;yy++){u32 *row=bb()+(u32)(y+yy)*fb_w+(u32)x;for(int xx=0;xx<w;xx++)row[xx]=c;}
}
static void blend_bb(int x,int y,int w,int h,u32 c){
  if(w<=0||h<=0)return;
  if(x<0){w+=x;x=0;}if(y<0){h+=y;y=0;}
  if((u32)(x+w)>fb_w)w=(int)fb_w-x;if((u32)(y+h)>fb_h)h=(int)fb_h-y;
  for(int yy=0;yy<h;yy++)for(int xx=0;xx<w;xx++){
    u32 *p=bb()+(u32)(y+yy)*fb_w+(u32)(x+xx);u32 d=*p;
    u32 r=(((d>>16)&0xFF)+((c>>16)&0xFF))>>1;
    u32 g=(((d>>8)&0xFF)+((c>>8)&0xFF))>>1;
    u32 b=((d&0xFF)+(c&0xFF))>>1;
    *p=0xFF000000u|(r<<16)|(g<<8)|b;
  }
}
static void rect_bb(int x,int y,int w,int h,u32 c){fill_bb(x,y,w,1,c);fill_bb(x,y+h-1,w,1,c);fill_bb(x,y,1,h,c);fill_bb(x+w-1,y,1,h,c);}
static void round_panel(int x,int y,int w,int h,u32 fill,u32 border){
  fill_bb(x+2,y,w-4,h,fill);fill_bb(x,y+2,w,h-4,fill);fill_bb(x+1,y+1,w-2,h-2,fill);
  rect_bb(x,y,w,h,border);
  put_bb(x,y,C_BG0);put_bb(x+w-1,y,C_BG0);put_bb(x,y+h-1,C_BG0);put_bb(x+w-1,y+h-1,C_BG0);
}
static void draw_char(int x,int y,char ch,u32 fg){
  int idx=(int)(u8)ch-32;if(idx<0||idx>=96)idx=0;const u8 *g=FONT8[idx];
  for(int row=0;row<8;row++){u8 bits=g[row];for(int col=0;col<8;col++)if(bits&(1<<col))
    for(int sy=0;sy<FONT_SCALE;sy++)for(int sx=0;sx<FONT_SCALE;sx++)put_bb(x+col*FONT_SCALE+sx,y+row*FONT_SCALE+sy,fg);}
}
static void draw_text(int x,int y,const char *s,u32 fg){while(*s){if((u32)x+CW>fb_w)break;draw_char(x,y,*s++,fg);x+=CW;}}
static void draw_text_clip(int x,int y,const char *s,u32 fg,int maxw){int cx=x;while(*s&&cx+CW<=x+maxw){draw_char(cx,y,*s++,fg);cx+=CW;}}
static void present(void){
  if(!fb_ok)return;
  for(u32 y=0;y<fb_h;y++){u32 *src=bb()+y*fb_w;u32 *dst=(u32*)((u8*)fb+y*fb_pitch);for(u32 x=0;x<fb_w;x++)dst[x]=src[x];}
}

static int ocb_add(const char *n,u32 kind,u32 caps){
  if(n_ocb>=MAX_OCB)return -1;if(buscar(n)>=0)return -2;
  struct OCB *o=&ocbs[n_ocb];
  o->magic=0x4F434200;o->oid=fnv(n)^nodo_id^(u32)n_ocb;o->kind=kind;o->caps=caps;
  o->pin=o->pout=0;o->alive=1;o->x=60+(n_ocb%5)*28;o->y=TOP_H+24+(n_ocb%4)*22;
  o->w=320;o->h=200;o->z=0;o->visible=1;o->focused=0;o->minimized=0;o->color=C_WIN;
  o->disp_clase=o->disp_dato=0;o->text[0]=0;o->text_len=0;o->n_items=0;o->sel_item=0;o->titulo[0]=0;
  cpy(o->nombre,n,MAX_NAME);return n_ocb++;
}
static int pulso(int from,int to){
  if(from<0||to<0)return -1;if(!tiene_cap(from,CAP_PULSO_ENV))return -2;if(!tiene_cap(to,CAP_PULSO_REC))return -3;
  ocbs[from].pout++;ocbs[to].pin++;nonce++;n_pulso++;return 0;
}
static int is_win(int i){
  if(i<0||i>=n_ocb||!ocbs[i].alive)return 0;
  u32 k=ocbs[i].kind;return k==K_WINDOW||k==K_SHELL||k==K_TEXT||k==K_LIST||k==K_ORGES||k==K_INSPECT;
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
  if(i>=0&&is_win(i)&&!ocbs[i].minimized){ocbs[i].focused=1;focus_win=i;z_raise(i);}else focus_win=-1;
}
static void text_append(int i,char ch){if(i<0||ocbs[i].kind!=K_TEXT)return;if(ocbs[i].text_len>=MAX_TEXT-1)return;ocbs[i].text[ocbs[i].text_len++]=ch;ocbs[i].text[ocbs[i].text_len]=0;}
static void text_backspace(int i){if(i<0||ocbs[i].kind!=K_TEXT||ocbs[i].text_len<=0)return;ocbs[i].text[--ocbs[i].text_len]=0;}
static int hit_win(int mx,int my){
  for(int zi=n_z-1;zi>=0;zi--){int i=z_order[zi];
    if(!ocbs[i].alive||!ocbs[i].visible||ocbs[i].minimized)continue;
    if(mx>=ocbs[i].x&&mx<ocbs[i].x+ocbs[i].w&&my>=ocbs[i].y&&my<ocbs[i].y+ocbs[i].h)return i;}
  return -1;
}

static int crear_orges(u32 kind,const char *base){
  char name[MAX_NAME];int n=0;while(base[n]&&n<10){name[n]=base[n];n++;}
  name[n++]='0'+(orges_seq%10);name[n]=0;orges_seq++;
  u32 caps=CAP_PULSO_ENV|CAP_PULSO_REC|CAP_DESKTOP;
  int i=ocb_add(name,kind,caps);if(i<0)return i;
  if(kind==K_TEXT){cpy(ocbs[i].titulo,"Campo de texto",MAX_TITLE);ocbs[i].w=360;ocbs[i].h=220;}
  else if(kind==K_LIST){cpy(ocbs[i].titulo,"Lista de items",MAX_TITLE);ocbs[i].n_items=4;
    cpy(ocbs[i].items[0],"Maestro",MAX_ITEM);cpy(ocbs[i].items[1],"Compositor",MAX_ITEM);
    cpy(ocbs[i].items[2],"Desktop",MAX_ITEM);cpy(ocbs[i].items[3],"Framebuffer",MAX_ITEM);ocbs[i].w=300;ocbs[i].h=240;}
  else if(kind==K_INSPECT){cpy(ocbs[i].titulo,"Inspector OCB",MAX_TITLE);ocbs[i].w=340;ocbs[i].h=280;}
  else cpy(ocbs[i].titulo,"Ventana orges",MAX_TITLE);
  z_add(i);focus_set(i);return i;
}
static void crear_from_dock(int kind_idx){
  if(kind_idx==0)crear_orges(K_WINDOW,"Win");
  else if(kind_idx==1)crear_orges(K_TEXT,"Txt");
  else if(kind_idx==2)crear_orges(K_LIST,"Lst");
  else if(kind_idx==3)crear_orges(K_INSPECT,"Ins");
  menu_open=0;
}

static void draw_wallpaper(void){
  for(u32 y=0;y<fb_h;y++){
    u32 t=(y*255)/(fb_h?fb_h:1);
    u32 r=7+(t*8)/255,g=11+(t*18)/255,b=20+(t*30)/255;
    u32 c=0xFF000000u|(r<<16)|(g<<8)|b;
    u32 *row=bb()+y*fb_w;
    for(u32 x=0;x<fb_w;x++){
      if((x%64)==0||(y%64)==0)row[x]=0xFF000000u|((r+4)<<16)|((g+6)<<8)|(b+8);
      else row[x]=c;
    }
  }
  int cx=(int)fb_w-120,cy=100,rad=70;
  for(int dy=-rad;dy<=rad;dy++)for(int dx=-rad;dx<=rad;dx++){
    int d2=dx*dx+dy*dy;if(d2>rad*rad)continue;
    u32 a=(u32)(rad*rad-d2)*40/(u32)(rad*rad);
    int px=cx+dx,py=cy+dy;if(px<0||py<0||(u32)px>=fb_w||(u32)py>=fb_h)continue;
    u32 *p=bb()+(u32)py*fb_w+(u32)px;u32 d=*p;
    u32 r=((d>>16)&0xFF)+a,g=((d>>8)&0xFF)+a*2,b=(d&0xFF)+a;
    if(r>255)r=255;if(g>255)g=255;if(b>255)b=255;
    *p=0xFF000000u|(r<<16)|(g<<8)|b;
  }
}
static void draw_topbar(void){
  fill_bb(0,0,(int)fb_w,TOP_H,C_TOP);fill_bb(0,TOP_H-1,(int)fb_w,1,C_BORDER);
  fill_bb(12,8,16,16,C_ACCENT);fill_bb(14,10,12,12,C_TOP);put_bb(18,14,C_ACCENT);
  draw_text(36,8,"AlsetOS",C_TEXT);draw_text(36+8*CW,8,"Desktop",C_DIM);
  char buf[48];int n=0;buf[n++]='O';buf[n++]=':';
  int alive=0;for(int i=0;i<n_ocb;i++)if(ocbs[i].alive)alive++;
  if(alive>=10){buf[n++]=(char)('0'+alive/10);buf[n++]=(char)('0'+alive%10);}else buf[n++]=(char)('0'+alive);
  buf[n++]=' ';buf[n++]='P';buf[n++]=':';
  if(n_pulso>=10){buf[n++]=(char)('0'+(n_pulso/10)%10);buf[n++]=(char)('0'+n_pulso%10);}else buf[n++]=(char)('0'+n_pulso);
  buf[n]=0;draw_text((int)fb_w-12-n*CW,8,buf,C_DIM);
}
static void draw_dock(void){
  int y=(int)fb_h-DOCK_H;fill_bb(0,y,(int)fb_w,DOCK_H,C_DOCK);fill_bb(0,y,(int)fb_w,1,C_BORDER);
  const char *labels[]={"Menu","Win","Text","List","Insp"};
  int nlab=5,slot=72,start=((int)fb_w-nlab*slot)/2;if(start<8)start=8;
  for(int i=0;i<nlab;i++){
    int bx=start+i*slot,by=y+8;
    round_panel(bx,by,56,36,C_ICON_BG,menu_open&&i==0?C_ACCENT:C_BORDER);
    draw_text(bx+6,by+10,labels[i],i==0&&menu_open?C_ACCENT:C_TEXT);
  }
}
static void draw_menu(void){
  if(!menu_open)return;
  int mw=240,mh=180,mx=16,my=(int)fb_h-DOCK_H-mh-8;
  blend_bb(mx+4,my+4,mw,mh,C_SHADOW);round_panel(mx,my,mw,mh,C_PANEL,C_ACCENT);
  draw_text(mx+12,my+12,"Crear organismo",C_ACCENT);
  draw_text(mx+12,my+40,"1  Ventana",C_TEXT);
  draw_text(mx+12,my+60,"2  Texto",C_TEXT);
  draw_text(mx+12,my+80,"3  Lista",C_TEXT);
  draw_text(mx+12,my+100,"4  Inspector",C_TEXT);
  draw_text(mx+12,my+128,"M cierra  Enter pulso",C_DIM);
  draw_text(mx+12,my+152,"Arrastra titulo",C_MUTED);
}
static void draw_window_body(int i){
  struct OCB *o=&ocbs[i];int x=o->x,y=o->y,w=o->w,h=o->h;
  blend_bb(x+6,y+8,w,h,C_SHADOW);
  u32 tb=o->focused?C_TITLE_F:C_TITLE;u32 bd=o->focused?C_ACCENT:C_BORDER;
  round_panel(x,y,w,h,C_WIN,bd);fill_bb(x+1,y+1,w-2,TITLE_H,tb);fill_bb(x+1,y+TITLE_H,w-2,1,bd);
  fill_bb(x+w-22,y+8,12,12,C_DANGER);fill_bb(x+w-40,y+8,12,12,C_WARN);fill_bb(x+w-58,y+8,12,12,C_ACCENT);
  draw_text_clip(x+10,y+6,o->titulo[0]?o->titulo:o->nombre,C_TEXT,w-70);
  int cy=y+TITLE_H+8,cx=x+12,cw=w-24;
  if(o->kind==K_TEXT){
    draw_text(cx,cy,"Editor de campo",C_DIM);
    fill_bb(cx,cy+CH+6,cw,h-TITLE_H-CH-24,C_PANEL2);rect_bb(cx,cy+CH+6,cw,h-TITLE_H-CH-24,C_BORDER);
    draw_text_clip(cx+6,cy+CH+12,o->text,C_TEXT,cw-12);
    if(o->focused&&((frame_tick/20)&1)){int tx=cx+6+o->text_len*CW;fill_bb(tx,cy+CH+12,2,CH,C_ACCENT);}
  }else if(o->kind==K_LIST){
    draw_text(cx,cy,"Lista organismo",C_DIM);
    for(int k=0;k<o->n_items;k++){
      int iy=cy+CH+8+k*(CH+6);if(iy+CH>y+h-8)break;
      u32 bg=(k==o->sel_item)?C_SEL:C_PANEL2;fill_bb(cx,iy-2,cw,CH+4,bg);
      if(k==o->sel_item)rect_bb(cx,iy-2,cw,CH+4,C_ACCENT);
      draw_text_clip(cx+8,iy,o->items[k],C_TEXT,cw-16);
    }
  }else if(o->kind==K_INSPECT){
    draw_text(cx,cy,"Inspector de red",C_DIM);
    char line[40];int row=0;
    for(int j=0;j<n_ocb&&row<8;j++){
      if(!ocbs[j].alive)continue;
      int n=0;const char *nm=ocbs[j].nombre;while(*nm&&n<12)line[n++]=*nm++;
      line[n++]=' ';line[n++]='#';u32 id=ocbs[j].oid&0xFF;const char *hex="0123456789ABCDEF";
      line[n++]=hex[(id>>4)&0xF];line[n++]=hex[id&0xF];line[n]=0;
      draw_text(cx,cy+CH+6+row*(CH+4),line,ocbs[j].focused?C_ACCENT:C_TEXT);row++;
    }
  }else{
    draw_text(cx,cy,"Organismo ventana",C_DIM);
    draw_text(cx,cy+CH+8,"Caps y pulsos activos.",C_TEXT);
    draw_text(cx,cy+2*(CH+8),"Arrastra la barra de titulo.",C_MUTED);
    int ox=cx+20,oy=cy+3*(CH+10);
    fill_bb(ox,oy,40,40,C_ICON_BG);rect_bb(ox,oy,40,40,C_ACCENT);
    fill_bb(ox+50,oy+14,30,4,C_ACCENT2);fill_bb(ox+90,oy,40,40,C_PANEL2);rect_bb(ox+90,oy,40,40,C_ACCENT2);
  }
}
static void draw_cursor(void){
  int x=mouse_x,y=mouse_y;
  for(int i=0;i<16;i++){put_bb(x,y+i,C_CURSOR);if(i<10)put_bb(x+i,y+i,C_CURSOR);}
  for(int i=0;i<8;i++)put_bb(x+1,y+1+i,C_ACCENT);
  rect_bb(x-2,y-2,6,6,C_ACCENT2);
}
static void compositor_tick(void){
  if(!fb_ok)return;frame_tick++;
  draw_wallpaper();draw_topbar();
  for(int zi=0;zi<n_z;zi++){int i=z_order[zi];if(ocbs[i].alive&&ocbs[i].visible&&!ocbs[i].minimized)draw_window_body(i);}
  draw_dock();draw_menu();draw_cursor();present();
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
  if(dx>2||dx<-2)dx*=2;if(dy>2||dy<-2)dy*=2;
  mouse_x+=dx;mouse_y-=dy;
  if(mouse_x<0)mouse_x=0;if(mouse_y<0)mouse_y=0;
  if((u32)mouse_x>=fb_w)mouse_x=(int)fb_w-1;if((u32)mouse_y>=fb_h)mouse_y=(int)fb_h-1;
  u8 prev=mouse_btn;mouse_btn=b0&0x07;
  if((mouse_btn&1)&&!(prev&1)){
    int dock_y=(int)fb_h-DOCK_H;
    if(mouse_y>=dock_y){
      int slot=72,nlab=5,start=((int)fb_w-nlab*slot)/2;if(start<8)start=8;
      int idx=(mouse_x-start)/slot;
      if(idx==0)menu_open=!menu_open;
      else if(idx>=1&&idx<=4)crear_from_dock(idx-1);
    }else{
      int hi=hit_win(mouse_x,mouse_y);
      if(hi>=0){
        focus_set(hi);
        if(mouse_y<ocbs[hi].y+TITLE_H){
          if(mouse_x>=ocbs[hi].x+ocbs[hi].w-22){
            ocbs[hi].alive=0;ocbs[hi].visible=0;
            int k=0;for(int j=0;j<n_z;j++)if(z_order[j]!=hi)z_order[k++]=z_order[j];n_z=k;focus_win=-1;
          }else if(mouse_x>=ocbs[hi].x+ocbs[hi].w-40){ocbs[hi].minimized=1;focus_win=-1;}
          else{drag_win=hi;drag_ox=mouse_x-ocbs[hi].x;drag_oy=mouse_y-ocbs[hi].y;}
        }
      }else menu_open=0;
    }
  }
  if(!(mouse_btn&1))drag_win=-1;
  if(drag_win>=0&&ocbs[drag_win].alive){
    ocbs[drag_win].x=mouse_x-drag_ox;ocbs[drag_win].y=mouse_y-drag_oy;
    if(ocbs[drag_win].y<TOP_H)ocbs[drag_win].y=TOP_H;
    if(ocbs[drag_win].y>(int)fb_h-DOCK_H-40)ocbs[drag_win].y=(int)fb_h-DOCK_H-40;
    if(ocbs[drag_win].x<0)ocbs[drag_win].x=0;
    if(ocbs[drag_win].x>(int)fb_w-40)ocbs[drag_win].x=(int)fb_w-40;
  }
  compositor_tick();
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
static void bootstrap_organisms(void){
  nodo_id=0xA15E7001u;
  ocb_add("Maestro",K_MASTER,CAP_ALL);
  ocb_add("Framebuffer",K_FB,CAP_FB|CAP_DISP_LEER|CAP_DISP_ESCR);
  ocb_add("Compositor",K_COMPOSE,CAP_COMPOSE|CAP_PULSO_ENV|CAP_PULSO_REC);
  ocb_add("Desktop",K_DESKTOP,CAP_DESKTOP|CAP_PULSO_ENV|CAP_PULSO_REC|CAP_ORGES_CREAR);
  ocb_add("Bus",K_BUS,CAP_PULSO_ENV|CAP_PULSO_REC);
  ocb_add("Teclado",K_DISP,CAP_DISP_LEER|CAP_DISP_CTRL);
  ocb_add("Raton",K_DISP,CAP_DISP_LEER|CAP_DISP_CTRL);
  int w=crear_orges(K_WINDOW,"Home");
  if(w>=0){cpy(ocbs[w].titulo,"Bienvenido a AlsetOS",MAX_TITLE);ocbs[w].x=80;ocbs[w].y=60;ocbs[w].w=420;ocbs[w].h=260;}
  int t=crear_orges(K_TEXT,"Note");
  if(t>=0){ocbs[t].x=200;ocbs[t].y=140;cpy(ocbs[t].text,"Escribe aqui...",MAX_TEXT);ocbs[t].text_len=14;}
  int L=crear_orges(K_LIST,"Net");
  if(L>=0){ocbs[L].x=500;ocbs[L].y=90;}
}

void kernel_main(u32 magic,u32 info){
  (void)info;
  for(int i=0;i<80*25;i++)VGA[i]=0x0F00|' ';
  vga_msg("AlsetOS: iniciando organismos...");
  if(!fb_init(magic,info)){vga_msg("AlsetOS: sin framebuffer");for(;;)pausa();}
  mouse_x=(int)fb_w/2;mouse_y=(int)fb_h/2;
  bootstrap_organisms();
  {int fi=buscar("Framebuffer");if(fi>=0){ocbs[fi].w=(int)fb_w;ocbs[fi].h=(int)fb_h;}}
  mouse_init();compositor_tick();
  for(;;){
    if(!(inb(0x64)&1)){
      if(focus_win>=0&&ocbs[focus_win].kind==K_TEXT&&(frame_tick%30)==0)compositor_tick();
      else pausa();
      frame_tick++;continue;
    }
    u8 st=inb(0x64);u8 sc=inb(0x60);
    if(st&0x20){mouse_feed(sc);continue;}
    if(sc==0xE0||(sc&0x80))continue;
    int ki=buscar("Teclado");if(ki>=0)ocbs[ki].disp_dato++;
    if(sc==0x32){menu_open=!menu_open;compositor_tick();continue;}
    if(sc==0x02){crear_orges(K_WINDOW,"Win");menu_open=0;compositor_tick();continue;}
    if(sc==0x03){crear_orges(K_TEXT,"Txt");menu_open=0;compositor_tick();continue;}
    if(sc==0x04){crear_orges(K_LIST,"Lst");menu_open=0;compositor_tick();continue;}
    if(sc==0x05){crear_orges(K_INSPECT,"Ins");menu_open=0;compositor_tick();continue;}
    if(sc==0x1C){int from=buscar("Desktop");if(from>=0&&focus_win>=0)pulso(from,focus_win);compositor_tick();continue;}
    if(sc==0x0F){int start=focus_win+1;for(int k=0;k<n_ocb;k++){int i=(start+k)%n_ocb;if(is_win(i)&&!ocbs[i].minimized){focus_set(i);break;}}compositor_tick();continue;}
    if(sc==0x0E){text_backspace(focus_win);compositor_tick();continue;}
    if(focus_win>=0&&ocbs[focus_win].kind==K_LIST){
      if(sc==0x48&&ocbs[focus_win].sel_item>0){ocbs[focus_win].sel_item--;compositor_tick();continue;}
      if(sc==0x50&&ocbs[focus_win].sel_item<ocbs[focus_win].n_items-1){ocbs[focus_win].sel_item++;compositor_tick();continue;}
    }
    if(focus_win>=0&&is_win(focus_win)){
      if(sc==0x4B){ocbs[focus_win].x-=24;compositor_tick();continue;}
      if(sc==0x4D){ocbs[focus_win].x+=24;compositor_tick();continue;}
      if(sc==0x48&&ocbs[focus_win].kind!=K_LIST){ocbs[focus_win].y-=24;compositor_tick();continue;}
      if(sc==0x50&&ocbs[focus_win].kind!=K_LIST){ocbs[focus_win].y+=24;compositor_tick();continue;}
    }
    char ch=sc_ascii(sc);
    if(ch&&focus_win>=0&&ocbs[focus_win].kind==K_TEXT){text_append(focus_win,ch);compositor_tick();}
  }
}
