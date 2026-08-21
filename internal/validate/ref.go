package validate

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/LYH2263/go-gitmirror/internal/errs"
)

// RefName 校验完整 ref 名（如 refs/heads/main）。
func RefName(name string) error {
	if name == "" || len(name) > 255 {
		return fmt.Errorf("%w: empty or too long", errs.ErrInvalidRef)
	}
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return fmt.Errorf("%w: leading/trailing slash", errs.ErrInvalidRef)
	}
	if strings.Contains(name, "//") || strings.Contains(name, "..") {
		return fmt.Errorf("%w: illegal sequence", errs.ErrInvalidRef)
	}
	if !strings.HasPrefix(name, "refs/") {
		return fmt.Errorf("%w: must start with refs/", errs.ErrInvalidRef)
	}
	for _, r := range name {
		if r == 0x7f || (r < 0x20 && r != '\t') {
			return fmt.Errorf("%w: control char", errs.ErrInvalidRef)
		}
		if unicode.IsSpace(r) && r != '-' {
			// 允许极少空白？Git 禁止空格
			if unicode.IsSpace(r) {
				return fmt.Errorf("%w: whitespace", errs.ErrInvalidRef)
			}
		}
	}
	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return fmt.Errorf("%w: too few components", errs.ErrInvalidRef)
	}
	return nil
}

// RefPattern 允许通配 refs/heads/* 。
func RefPattern(pat string) error {
	if strings.HasSuffix(pat, "/*") {
		base := strings.TrimSuffix(pat, "/*")
		if base == "refs/heads" || base == "refs/tags" || base == "refs/remotes" {
			return nil
		}
		if strings.HasPrefix(base, "refs/") && !strings.Contains(base, "*") {
			return nil
		}
	}
	return RefName(pat)
}

// ObjectID 校验 40 hex。
func ObjectID(oid string) error {
	if len(oid) != 40 {
		return errs.ErrBadOID
	}
	for _, c := range oid {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return errs.ErrBadOID
		}
	}
	return nil
}
