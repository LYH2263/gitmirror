package validate

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/LYH2263/go-gitmirror/internal/errs"
)

// RemoteURL 接受 https/http/git/ssh/file 与 memory:// 测试 scheme。
func RemoteURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("%w: empty url", errs.ErrInvalidRemote)
	}
	if strings.HasPrefix(raw, "memory://") {
		return nil
	}
	if strings.HasPrefix(raw, "git@") && strings.Contains(raw, ":") {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: %v", errs.ErrInvalidRemote, err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "git", "ssh", "file":
		if u.Host == "" && u.Scheme != "file" {
			return fmt.Errorf("%w: missing host", errs.ErrInvalidRemote)
		}
		return nil
	default:
		return fmt.Errorf("%w: unsupported scheme %q", errs.ErrInvalidRemote, u.Scheme)
	}
}

// HeaderKey 简单校验。
func HeaderKey(k string) error {
	if k == "" || strings.ContainsAny(k, " \r\n:") {
		return fmt.Errorf("%w: bad header key", errs.ErrInvalidRemote)
	}
	return nil
}
