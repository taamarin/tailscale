// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause
//go:build !android

package taildrop

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var renameMu sync.Mutex

func NewDefaultFileOps(root string) DefaultFileOps { return DefaultFileOps{rootDir: root} }

func (d DefaultFileOps) OpenWriter(name string, offset int64, perm os.FileMode) (io.WriteCloser, string, error) {
	path := filepath.Join(d.rootDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, "", err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return nil, "", err
	}
	if offset != 0 {
		curr, err := f.Seek(0, io.SeekEnd)
		if err != nil {
			f.Close()
			return nil, "", err
		}
		if offset < 0 || offset > curr {
			f.Close()
			return nil, "", fmt.Errorf("offset %d out of range", offset)
		}
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			f.Close()
			return nil, "", err
		}
		if err := f.Truncate(offset); err != nil {
			f.Close()
			return nil, "", err
		}
	}
	return f, path, nil
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

// Rename moves the partial file into its final name.
// If finalName contains any path separators (or is absolute),
// we use it verbatim; otherwise we join it to the partial’s dir.
// It will retry up to 10 times, de-dup same-checksum files, etc.
func (DefaultFileOps) Rename(partial, finalName string) (string, error) {
	var dst string
	if filepath.IsAbs(finalName) || strings.ContainsRune(finalName, os.PathSeparator) {
		dst = finalName
	} else {
		dir := filepath.Dir(partial)
		if base := filepath.Base(partial); strings.HasSuffix(base, partialSuffix) {
			trimmed := strings.TrimSuffix(base, partialSuffix)
			first := strings.IndexByte(trimmed, '.')
			last := strings.LastIndexByte(trimmed, '.')
			if first != -1 && last > first {
				id := trimmed[last+1:]
				dir = filepath.Join(dir, id)
			}
		}
		dst = filepath.Join(dir, finalName)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}

	st, err := os.Stat(partial)
	if err != nil {
		return "", err
	}
	wantSize := st.Size()

	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		renameMu.Lock()
		fi, statErr := os.Stat(dst)
		// Atomically rename the partial file as the destination file if it doesn't exist.
		// Otherwise, it returns the length of the current destination file.
		// The operation is atomic.
		if os.IsNotExist(statErr) {
			err = os.Rename(partial, dst)
			renameMu.Unlock()
			if err != nil {
				return "", err
			}
			return dst, nil
		}
		if statErr != nil {
			renameMu.Unlock()
			return "", statErr
		}
		gotSize := fi.Size()
		renameMu.Unlock()

		// Avoid the final rename if a destination file has the same contents.
		//
		// Note: this is best effort and copying files from iOS from the Media Library
		// results in processing on the iOS side which means the size and shas of the
		// same file can be different.
		if gotSize == wantSize {
			sumP, err := sha256File(partial)
			if err != nil {
				return "", err
			}
			sumD, err := sha256File(dst)
			if err != nil {
				return "", err
			}
			if bytes.Equal(sumP[:], sumD[:]) {
				if err := os.Remove(partial); err != nil {
					return "", err
				}
				return dst, nil
			}
		}

		// Choose a new destination filename and try again.
		dst = filepath.Join(filepath.Dir(dst), nextFilename(filepath.Base(dst)))
	}

	return "", fmt.Errorf("too many retries trying to rename %q to %q", partial, finalName)
}

// sha256File computes the SHA‑256 of a file.
func sha256File(path string) ([sha256.Size]byte, error) {
	var sum [sha256.Size]byte
	f, err := os.Open(path)
	if err != nil {
		return sum, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return sum, err
	}
	copy(sum[:], h.Sum(nil))
	return sum, nil
}

func (DefaultFileOps) ListDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}
