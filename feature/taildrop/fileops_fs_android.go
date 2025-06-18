// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

//go:build android

package taildrop

import (
	"io"
	"os"
)

// defaultFileOps starts nil on Android; libtailscale will replace it
// with an AndroidFileOps during backend construction.
var defaultFileOps FileOps = nil

// NewDefaultFileOps satisfies references from ext.go but is basically a
// no‑op on Android. We still return a value of the proper type so the call
// site compiles, but it won’t be used because libtailscale will install an
// AndroidFileOps immediately after backend creation.
func NewDefaultFileOps(root string) DefaultFileOps {
	return DefaultFileOps{rootDir: root}
}

func (d DefaultFileOps) OpenWriter(name string, offset int64, perm os.FileMode) (io.WriteCloser, string, error) {
	return nil, "", os.ErrPermission
}

func (d DefaultFileOps) Base(pathOrURI string) string              { return pathOrURI }
func (d DefaultFileOps) Remove(name string) error                  { return os.ErrPermission }
func (d DefaultFileOps) ListDir(dir string) ([]os.DirEntry, error) { return nil, os.ErrPermission }
func (d DefaultFileOps) Rename(oldPathOrURI, newName string) (string, error) {
	return "", os.ErrPermission
}
