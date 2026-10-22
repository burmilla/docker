//go:build !windows
// +build !windows

package symlink

import (
	"path/filepath"
)

func isDriveOrRoot(p string) bool {
	return p == string(filepath.Separator)
}
