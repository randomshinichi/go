	.text
	.globl _probe
_probe:
	bl	_ext
	lis	r3, ha16(_global+8)
	addi	r3, r3, lo16(_global+8)
	lis	r4, ha16(_global+8)
	ld	r4, lo16(_global+8)(r4)
	blr
