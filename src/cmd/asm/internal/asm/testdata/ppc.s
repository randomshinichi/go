#include "../../../../../runtime/textflag.h"
TEXT ·probe(SB),NOSPLIT,$0-0 // TEXT pkg.probe(SB), NOSPLIT, $0-0
	LWARX R3,R4,R5 // LWARX R4, R5, R3 // 7c642828
	STWCX R3,R4,R5 // STWCX. R3, R5, R4 // 7c64292d
	SYNC // SYNC // 7c0004ac
	ISYNC // ISYNC // 4c00012c
	RET
