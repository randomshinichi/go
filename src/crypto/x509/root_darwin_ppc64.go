// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package x509

// D18: Leopard lacks five modern Security.framework APIs (measured in
// out/stage2-20261003/parent-native-run/D18-FRAMEWORK-MEASUREMENTS.md).
// This target uses an updatable PEM CA bundle and Go's verifier, NOT Apple's
// keychain, user/admin distrust settings, or Security-framework verification.
// Trust is determined solely by the selected file/directory; users must manage
// distrust by removing a root from that source. No security(1) subprocess runs.
// The measured paths below are from
// out/stage2-20261003/parent-native-run/D18-ADDENDUM4-NETWORK-TOOLING-CA.md.
// Unlike root_unix.go, SSL_CERT_FILE suppresses ALL directory/default sources;
// otherwise SSL_CERT_DIR suppresses default files. An explicit file failure is
// a hard error. Listed directories that do not exist are skipped; any other
// error reading a listed directory is a hard error, even if another has roots.
// Unreadable entries within directories are skipped, as in the Unix loader;
// zero certificates overall is a hard error. No explicit failure falls back to
// defaults. Without either override, the first readable default file wins
// (no directory union).
// The pool is cached by root.go:systemRootsPool, as on other platforms.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var darwinPPC64CertFiles = []string{
	"/usr/local/etc/openssl/ca-bundle-current.crt",      // Mozilla 2026-09-25, placed by owner.
	"/usr/local/etc/openssl/cert.pem",                   // Homebrew Mozilla 2021-09-30 bundle.
	"/usr/local/opt/curl-ca-bundle/share/ca-bundle.crt", // Same bundle, explicit Homebrew path.
	"/usr/share/curl/curl-ca-bundle.crt",                // System 2007 bundle, last resort.
}

func (c *Certificate) systemVerify(opts *VerifyOptions) ([][]*Certificate, error) {
	return nil, errors.New("x509: Security-framework verification is unavailable on darwin/ppc64")
}

func loadSystemRoots() (*CertPool, error) {
	if file := os.Getenv("SSL_CERT_FILE"); file != "" {
		return loadDarwinPPC64CertFile(file)
	}
	if dirs := os.Getenv("SSL_CERT_DIR"); dirs != "" {
		roots := NewCertPool()
		var firstErr error
		for _, dir := range filepath.SplitList(dirs) {
			entries, err := os.ReadDir(dir)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("x509: cannot read SSL_CERT_DIR directory %q: %w", dir, err)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
				if err != nil {
					if firstErr == nil {
						firstErr = err
					}
					continue
				}
				// Hashed symlinks and their targets may both occur; CertPool
				// deduplicates certificates, as in the Unix loader.
				roots.AppendCertsFromPEM(data)
			}
		}
		if roots.len() == 0 {
			if firstErr != nil {
				return nil, fmt.Errorf("x509: SSL_CERT_DIR contains no certificates: %w", firstErr)
			}
			return nil, errors.New("x509: SSL_CERT_DIR contains no certificates")
		}
		return roots, nil
	}
	roots, _, err := loadDarwinPPC64DefaultRoots()
	return roots, err
}

// loadDarwinPPC64DefaultRoots returns the selected path with the pool so tests
// can distinguish the current bundle from the older, last-resort bundles.
func loadDarwinPPC64DefaultRoots() (*CertPool, string, error) {
	var firstErr error
	for _, file := range darwinPPC64CertFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		// A readable but empty/invalid bundle is an error, not permission to
		// trust a different source. This is the first-readable rule.
		roots, err := darwinPPC64CertPool(file, data)
		return roots, file, err
	}
	return nil, "", fmt.Errorf("x509: no readable CA bundle: %w", firstErr)
}

func loadDarwinPPC64CertFile(file string) (*CertPool, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("x509: cannot read CA bundle %q: %w", file, err)
	}
	return darwinPPC64CertPool(file, data)
}

func darwinPPC64CertPool(file string, data []byte) (*CertPool, error) {
	roots := NewCertPool()
	if !roots.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("x509: CA bundle %q contains no certificates", file)
	}
	return roots, nil
}
