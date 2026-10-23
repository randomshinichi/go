#!/bin/bash
# gen-target.sh — run ON Mac OS X 10.5 PowerPC with a native darwin/ppc Go.
# usage: gen-target.sh KITDIR   (GOROOT and PATH must select the native toolchain)
set -u
cd "$1" || exit 2
export GOOS=darwin GOARCH=ppc CGO_ENABLED=1 GOENV=off GOFLAGS= GO111MODULE=off GOTOOLCHAIN=local
export CC="$PWD/cc-ppc.sh"
{ uname -a; sw_vers; /usr/bin/gcc --version 2>&1 | head -1; command -v go; go version; } > host-identity.txt 2>&1
bash ./mkerrors.sh > zerrors.raw 2> zerrors.err; echo "mkerrors rc=$?" >> rc.txt
go tool cgo -godefs types_darwin_leopard.go > ztypes.raw 2> ztypes.err; echo "godefs rc=$?" >> rc.txt
"$CC" -o xsys_sizes xsys_sizes.c 2> sizes.err && ./xsys_sizes > xsys_sizes.txt; echo "sizes rc=$?" >> rc.txt
rm -rf xsys_sizes _obj _const.go _error.grep _signal.grep _error.out _errors.c _errors
