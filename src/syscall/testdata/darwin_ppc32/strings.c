#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <signal.h>
static int ec[] = {7,13,48,49,47,35,37,80,86,85,9,88,94,72,16,89,10,53,61,54,11,39,83,33,69,17,14,27,79,64,65,90,92,36,4,22,5,56,21,103,62,24,31,40,95,63,81,50,52,51,23,93,55,96,19,2,8,77,97,12,91,103,42,28,98,99,78,15,57,20,66,38,45,25,6,102,84,1,46,32,67,76,75,74,100,43,41,82,34,71,30,73,87,58,44,29,3,70,101,60,59,26,68,35,18};
static int sc[] = {6,14,10,20,19,7,8,1,4,29,2,23,6,9,13,27,3,11,17,12,15,5,18,21,22,16,30,31,26,28,24,25};
static int cmp(const void *a,const void *b){return *(int*)a-*(int*)b;}
int main(void){int i,e; char buf[1024], *p; qsort(ec,sizeof(ec)/sizeof(ec[0]),sizeof(int),cmp); puts("ERRORS"); for(i=0;i<sizeof(ec)/sizeof(ec[0]);i++){e=ec[i];if(i&&e==ec[i-1])continue;strncpy(buf,strerror(e),sizeof(buf)-1);buf[sizeof(buf)-1]=0;p=buf;if(p[0]>='A'&&p[0]<='Z'&&p[1]>='a'&&p[1]<='z')p[0]+=32;printf("%d\t%s\n",e,p);} qsort(sc,sizeof(sc)/sizeof(sc[0]),sizeof(int),cmp); puts("SIGNALS"); for(i=0;i<sizeof(sc)/sizeof(sc[0]);i++){e=sc[i];if(i&&e==sc[i-1])continue;strncpy(buf,strsignal(e),sizeof(buf)-1);buf[sizeof(buf)-1]=0;p=buf;if(p[0]>='A'&&p[0]<='Z'&&p[1]>='a'&&p[1]<='z')p[0]+=32;char *q=strrchr(p,':');if(q)*q=0;printf("%d\t%s\n",e,p);}return 0;}
