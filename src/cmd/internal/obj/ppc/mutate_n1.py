# F2a: DROP the opening Restart1 marker. The bug is live again.
L = "/data/source/projects/go-g5-port/go/worktrees/ppc32-r31nf/src/cmd/internal/obj/ppc/link.go"
s = open(L).read()
old = "\t\tp = appendUnsafePoint(p, abi.UnsafePointRestart1)\n"
assert s.count(old) == 1, s.count(old)
s = s.replace(old, "\t\t_ = abi.UnsafePointRestart1\n\t\t_ = restore\n")
open(L, "w").write(s)
