// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

//go:build !android

package taildrop

import (
	"io"
	"os"
	"path/filepath"
)

// DefaultFileOps is the non-Android implementation of FileOps.
type DefaultFileOps struct{}

func (DefaultFileOps) OpenWriter(dest string, offset int64, perm os.FileMode) (io.WriteCloser, string, error) {
	partial := dest + ".part"
	f, err := os.OpenFile(partial, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return nil, "", err
	}
	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			f.Close()
			return nil, "", err
		}
		if err := f.Truncate(offset); err != nil {
			f.Close()
			return nil, "", err
		}
	}
	return f, partial, nil
}

func (DefaultFileOps) Base(pathOrURI string) string {
	return filepath.Base(pathOrURI)
}

func (DefaultFileOps) Remove(name string) error {
	return os.Remove(name)
}

func (DefaultFileOps) Join(dir, name string) string {
	return filepath.Join(dir, name)
}

func (DefaultFileOps) Rename(partial, finalName string) (string, error) {
	if err := os.Rename(partial, finalName); err != nil {
		return "", err
	}
	return finalName, nil
}
