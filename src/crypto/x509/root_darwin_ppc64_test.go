// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package x509

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBundleRootsDarwinPPC64(t *testing.T) {
	dir := t.TempDir()
	fileA, fileB := filepath.Join(dir, "a.pem"), filepath.Join(dir, "b.pem")
	for file, data := range map[string]string{fileA: gtsRoot, fileB: digicertRoot} {
		if err := os.WriteFile(file, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	parse := func(data string) *Certificate {
		t.Helper()
		block, _ := pem.Decode([]byte(data))
		cert, err := ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatal(err)
		}
		return cert
	}
	a, b := parse(gtsRoot), parse(digicertRoot)
	oldFiles := darwinPPC64CertFiles
	t.Cleanup(func() { darwinPPC64CertFiles = oldFiles })
	darwinPPC64CertFiles = []string{filepath.Join(dir, "missing"), fileB, fileA}
	assertPool := func(t *testing.T, roots *CertPool, err error, want, reject *Certificate) {
		t.Helper()
		if err != nil || roots == nil {
			t.Fatalf("load: %v", err)
		}
		if roots.systemPool || roots.len() != 1 || !roots.contains(want) || roots.contains(reject) {
			t.Fatal("pool did not contain exactly the selected root")
		}
		if _, err := want.Verify(VerifyOptions{Roots: roots, CurrentTime: want.NotBefore.Add(time.Hour), KeyUsages: []ExtKeyUsage{ExtKeyUsageAny}}); err != nil {
			t.Fatalf("selected root rejected: %v", err)
		}
		_, err = reject.Verify(VerifyOptions{Roots: roots, CurrentTime: reject.NotBefore.Add(time.Hour), KeyUsages: []ExtKeyUsage{ExtKeyUsageAny}})
		var unknown UnknownAuthorityError
		if !errors.As(err, &unknown) {
			t.Fatalf("unchosen root must be rejected as UnknownAuthorityError, got %v", err)
		}
	}
	t.Run("file-only-no-directory-or-default-union", func(t *testing.T) {
		t.Setenv("SSL_CERT_FILE", fileA)
		t.Setenv("SSL_CERT_DIR", dir) // contains the other, explicitly excluded root
		roots, err := loadSystemRoots()
		assertPool(t, roots, err, a, b)
	})
	t.Run("explicit-file-errors-do-not-fallback", func(t *testing.T) {
		invalid := filepath.Join(dir, "invalid.pem")
		if err := os.WriteFile(invalid, []byte("not a certificate"), 0600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("SSL_CERT_DIR", dir)
		for _, file := range []string{filepath.Join(dir, "absent"), invalid, dir} {
			t.Setenv("SSL_CERT_FILE", file)
			if roots, err := loadSystemRoots(); err == nil || roots != nil {
				t.Fatalf("%q: got pool %v, err %v; want hard error", file, roots, err)
			}
		}
	})
	t.Run("directory-only-hashed-symlink-deduplicated", func(t *testing.T) {
		chosen := t.TempDir()
		if err := os.WriteFile(filepath.Join(chosen, "root.pem"), []byte(gtsRoot), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("root.pem", filepath.Join(chosen, "01234567.0")); err != nil {
			t.Fatal(err)
		}
		t.Setenv("SSL_CERT_FILE", "")
		t.Setenv("SSL_CERT_DIR", t.TempDir()+string(os.PathListSeparator)+chosen)
		roots, err := loadSystemRoots()
		assertPool(t, roots, err, a, b)
	})
	t.Run("empty-or-unreadable-directory-errors", func(t *testing.T) {
		t.Setenv("SSL_CERT_FILE", "")
		for _, dirs := range []string{t.TempDir(), filepath.Join(dir, "absent-dir")} {
			t.Setenv("SSL_CERT_DIR", dirs)
			if roots, err := loadSystemRoots(); err == nil || roots != nil {
				t.Fatalf("%q: want hard error, got %v, %v", dirs, roots, err)
			}
		}
	})
	t.Run("defaults-first-readable-only", func(t *testing.T) {
		t.Setenv("SSL_CERT_FILE", "")
		t.Setenv("SSL_CERT_DIR", "")
		roots, err := loadSystemRoots()
		assertPool(t, roots, err, b, a)
	})
	t.Run("readable-invalid-default-stops-search", func(t *testing.T) {
		t.Setenv("SSL_CERT_FILE", "")
		t.Setenv("SSL_CERT_DIR", "")
		bad := filepath.Join(dir, "bad-default")
		if err := os.WriteFile(bad, nil, 0600); err != nil {
			t.Fatal(err)
		}
		darwinPPC64CertFiles = []string{bad, fileB}
		if roots, err := loadSystemRoots(); roots != nil || err == nil {
			t.Fatalf("invalid readable default: %v, %v", roots, err)
		}
	})
	t.Run("no-readable-default-errors", func(t *testing.T) {
		t.Setenv("SSL_CERT_FILE", "")
		t.Setenv("SSL_CERT_DIR", "")
		darwinPPC64CertFiles = []string{filepath.Join(dir, "missing-default")}
		if roots, err := loadSystemRoots(); roots != nil || err == nil {
			t.Fatalf("missing default: %v, %v", roots, err)
		}
	})
}

func TestBundleRootFailureDispatchDarwinPPC64(t *testing.T) {
	// Initialize the once before overriding globals, following root_test.go.
	systemRootsPool()
	oldRoots, oldErr := systemRoots, systemRootsErr
	t.Cleanup(func() { systemRoots, systemRootsErr = oldRoots, oldErr })
	marker := errors.New("D18 intentional root-loading failure")
	systemRoots, systemRootsErr = nil, marker
	block, _ := pem.Decode([]byte(gtsRoot))
	cert, err := ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	chains, err := cert.Verify(VerifyOptions{CurrentTime: cert.NotBefore.Add(time.Hour)})
	var rootsErr SystemRootsError
	if len(chains) != 0 || !errors.As(err, &rootsErr) || !errors.Is(err, marker) {
		t.Fatalf("root-load error did not propagate through Go verifier: chains=%v, err=%v", chains, err)
	}
	if chains, err := cert.systemVerify(&VerifyOptions{}); len(chains) != 0 || err == nil {
		t.Fatalf("unavailable platform verifier must fail loudly: chains=%v, err=%v", chains, err)
	}
}

func TestSystemBundleDarwinPPC64(t *testing.T) {
	t.Setenv("SSL_CERT_FILE", "")
	t.Setenv("SSL_CERT_DIR", "")
	roots, err := loadSystemRoots()
	if err != nil || roots == nil || roots.len() == 0 || roots.systemPool {
		t.Fatalf("system bundle: %v", err)
	}
	t.Logf("loaded %d roots from the first readable configured bundle", roots.len())
	root, _, err := roots.cert(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := root.Verify(VerifyOptions{Roots: roots, CurrentTime: root.NotBefore.Add(time.Hour), KeyUsages: []ExtKeyUsage{ExtKeyUsageAny}}); err != nil {
		t.Fatalf("selected bundle root rejected: %v", err)
	}
	key := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	template := &Certificate{
		SerialNumber: big.NewInt(18), Subject: pkix.Name{CommonName: "D18 isolated negative root"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: KeyUsageCertSign,
	}
	der, err := CreateCertificate(rand.Reader, template, template, key.Public(), key)
	if err != nil {
		t.Fatal(err)
	}
	unknown, err := ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	if roots.contains(unknown) {
		t.Fatal("negative root unexpectedly present in chosen bundle")
	}
	_, err = unknown.Verify(VerifyOptions{Roots: roots, KeyUsages: []ExtKeyUsage{ExtKeyUsageAny}})
	var unknownErr UnknownAuthorityError
	if !errors.As(err, &unknownErr) {
		t.Fatalf("unchosen root accepted or rejected for wrong reason: %v", err)
	}
}
