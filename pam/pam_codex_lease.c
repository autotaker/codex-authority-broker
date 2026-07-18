#define _GNU_SOURCE
#include <security/pam_modules.h>
#include <stdint.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>
#define MAGIC 0x43415831U
static void put32(unsigned char*p,uint32_t v){p[0]=v>>24;p[1]=v>>16;p[2]=v>>8;p[3]=v;}
static uint32_t get32(unsigned char*p){return((uint32_t)p[0]<<24)|((uint32_t)p[1]<<16)|((uint32_t)p[2]<<8)|p[3];}
static int live(void){int f=socket(AF_UNIX,SOCK_SEQPACKET|SOCK_CLOEXEC,0),ok=0;struct sockaddr_un a;unsigned char q[12]={0},r[268];ssize_t n;if(f<0)return 0;memset(&a,0,sizeof a);a.sun_family=AF_UNIX;strcpy(a.sun_path,"/run/codex-authority/authority.sock");put32(q,MAGIC);q[4]=1;q[5]=1;if(connect(f,(void*)&a,sizeof a)||send(f,q,12,MSG_NOSIGNAL)!=12)goto out;n=recv(f,r,sizeof r,0);if(n==12&&get32(r)==MAGIC&&r[4]==1&&r[5]==1&&r[6]==0&&get32(r+8)==0)ok=1;out:close(f);return ok;}
PAM_EXTERN int pam_sm_authenticate(pam_handle_t*p,int f,int c,const char**v){const char*u=0;(void)f;(void)c;(void)v;if(pam_get_user(p,&u,0)!=PAM_SUCCESS||!u||strcmp(u,"codex"))return PAM_AUTH_ERR;return live()?PAM_SUCCESS:PAM_AUTH_ERR;}
PAM_EXTERN int pam_sm_setcred(pam_handle_t*p,int f,int c,const char**v){(void)p;(void)f;(void)c;(void)v;return PAM_SUCCESS;}
