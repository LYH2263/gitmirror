package object

import (
	"path/filepath"
	"strings"
)

// LoosePath 返回 objects/ab/cdef... 松散对象路径。
func LoosePath(root string, oid string) string {
	oid = strings.ToLower(oid)
	if len(oid) < 3 {
		return filepath.Join(root, "objects", oid)
	}
	return filepath.Join(root, "objects", oid[:2], oid[2:])
}

// PackDir objects/pack。
func PackDir(root string) string {
	return filepath.Join(root, "objects", "pack")
}
