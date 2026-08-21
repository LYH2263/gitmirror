package object

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/LYH2263/go-gitmirror/internal/errs"
)

// ID 类型化 OID。
type ID [20]byte

// Parse 解析 40 hex。
func Parse(s string) (ID, error) {
	var id ID
	if len(s) != 40 {
		return id, errs.ErrBadOID
	}
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 20 {
		return id, errs.ErrBadOID
	}
	copy(id[:], b)
	return id, nil
}

// String 小写 hex。
func (id ID) String() string {
	return hex.EncodeToString(id[:])
}

// Zero 是否全 0。
func (id ID) Zero() bool {
	for _, b := range id {
		if b != 0 {
			return false
		}
	}
	return true
}

// HashBlob 计算 git blob OID（含 header）。
func HashBlob(data []byte) ID {
	h := sha1.New()
	_, _ = fmt.Fprintf(h, "blob %d\x00", len(data))
	_, _ = h.Write(data)
	var id ID
	copy(id[:], h.Sum(nil))
	return id
}

// IsHexOID 快速判断。
func IsHexOID(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, c := range strings.ToLower(s) {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
