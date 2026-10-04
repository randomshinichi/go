	.text
	.globl _probe
_probe:
	lis	r3, ha16(_g+0x18000)
	addi	r3, r3, lo16(_g+0x18000)
	lis	r4, ha16(_g-8)
	addi	r4, r4, lo16(_g-8)
	lis	r5, ha16(_g+0x8000)
	addi	r5, r5, lo16(_g+0x8000)
	lis	r6, ha16(_g+0x18000)
	ld	r6, lo16(_g+0x18000)(r6)
	blr
