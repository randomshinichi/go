# `d17readdir` — regression test for darwin/ppc64 `readdir_r` over `readdir_r$INODE64`

A test-only module. It lives under `testdata/` because it needs its own
assembly file and its own dynamic import, and neither can be added to the
`syscall` package without changing every darwin/ppc64 image (see below).

## What it tests

Plain `readdir_r` on Leopard fills the **legacy 32-bit-inode** `struct dirent`
(264 bytes: `fileno@0(4) reclen@4(2) type@6(1) namlen@7(1) name@8`), while
`syscall.Dirent` describes the **64-bit-inode** record (1048 bytes:
`ino@0(8) seekoff@8(8) reclen@16(2) namlen@18(2) type@20(1) name@21`). The
mispairing is silent: the count is right, no error is returned, every field is
wrong. `syscall.readdir_r` therefore imports `readdir_r$INODE64`.

Nothing in the standard library calls `readdir_r` on darwin/ppc64 (`os`'s
libc-directory implementation and `syscall`'s `Getdirentries`-over-`readdir_r`
are `!ppc64`; `Getdirentries` there is a raw `SYS_getdirentries64`), so the only
reachable caller is a third party using the linkname that `syscall` pushes:

    src/syscall/linkname_darwin.go   //go:linkname readdir_r
    src/os/dir_darwin.go             //go:linkname readdir_r syscall.readdir_r

This module is that caller. `readdirr_darwin_ppc64_test.go` reaches the function
through `//go:linkname readdir_r syscall.readdir_r`, typed with the exported
`syscall.Dirent`, and compares every record it fills with

* the record the kernel returns through `syscall.Getdirentries` (raw
  `SYS_getdirentries64`) for the same directory, field by field, and
* the inode `stat64` reports for the same name.

## Why `testdata/` and not `src/syscall/*_test.go`

`readdir_r` needs a `DIR*`, and darwin/ppc64 has no `fdopendir` to hand one out
from a descriptor, so the test must open the directory itself with
`opendir$INODE64` (the 64-bit-inode member of the family, i.e. the `DIR` variant
`readdir_r$INODE64` expects; the two variants keep different `DIR` structures).
That needs a dynamic import and an assembly trampoline. In `src/syscall` the
import would be a new entry point on the shipped package for a call nothing
makes, and a `.s` file is part of the package's build (a file named `*_test.s`
is still package assembly: `go/build` treats the `_test` suffix as an OS/arch
suffix, not as test-only), so the trampoline would ship in every image. Here it
stays in a module the toolchain does not build as part of `std`.

## Running it

```sh
cd $GOROOT/src/syscall/testdata/d17readdir
env GOROOT=$GOROOT GOOS=darwin GOARCH=ppc64 CGO_ENABLED=0 GOPPC64=ppc970 \
    GOPROXY=off GOTOOLCHAIN=local go test -c -o d17-readdir-darwin-ppc64.test .
```

The binary runs on the G5 (`-test.v`). `records_host_test.go` and the comparison
in `records.go` also run on the host (`go test .` on linux/amd64): they fabricate
the bytes a legacy `readdir_r` and an `$INODE64` one write, and show that the
comparison accepts the latter and rejects the former, so the test's
discriminating power is demonstrated without a G5.

Sources and evidence for D17: `out/stage2-20261003/D17-readdir-inode64/`.