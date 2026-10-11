# F2b: END the window one real instruction PAST the STW (it also covers the
# ADDI that zeroes R0). Width stays legal at 12 bytes, so only closure catches it.
L = "/data/source/projects/go-g5-port/go/worktrees/ppc32-r31nf/src/cmd/internal/obj/ppc/link.go"
s = open(L).read()
old = ("\t\tp = appendUnsafePoint(p, restore)\n"
       "\t\t// Go treats R0 as zero for stores (not merely the RA=0 address rule).\n"
       "\t\tp = appendOp(p, AADDI, obj.Addr{Type: obj.TYPE_CONST}, reg(REG_R0))\n"
       "\t\tp.Reg = REG_R0\n")
assert s.count(old) == 1
s = s.replace(old, ("\t\t// Go treats R0 as zero for stores (not merely the RA=0 address rule).\n"
       "\t\tp = appendOp(p, AADDI, obj.Addr{Type: obj.TYPE_CONST}, reg(REG_R0))\n"
       "\t\tp.Reg = REG_R0\n"
       "\t\tp = appendUnsafePoint(p, restore)\n"))
open(L, "w").write(s)
