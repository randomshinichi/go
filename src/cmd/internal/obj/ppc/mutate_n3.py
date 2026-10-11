# F2c: WIDEN past the 20-byte limit: hold the Restart1 run open across the whole
# stack-guard branch and the frame allocation, then close it. runtime/preempt.go
# would throw "bad restart PC" on any preemption inside it.
L = "/data/source/projects/go-g5-port/go/worktrees/ppc32-r31nf/src/cmd/internal/obj/ppc/link.go"
s = open(L).read()
old = "\t\tp = appendUnsafePoint(p, restore)\n"
assert s.count(old) == 1
s = s.replace(old, "\t\t_ = restore\n")
anchor = "\tfor p := text.Link; p != nil; p = p.Link {\n\t\tif p.As == obj.AGETCALLERPC {"
assert s.count(anchor) == 1
s = s.replace(anchor, "\tp = appendUnsafePoint(p, int64(abi.UnsafePointSafe))\n" + anchor)
open(L, "w").write(s)
