package errs

import (
	"errors"
	"fmt"
	"testing"
)

func TestAuthErrorIsErrAuthFailed(t *testing.T) {
	// AsAuth 构造的认证错误必须被 errors.Is 识别为 ErrAuthFailed。
	// 历史问题：AuthError.Unwrap 返回 nil，导致该判断恒为 false。
	err := AsAuth("token mismatch")
	if !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("errors.Is(AsAuth(...), ErrAuthFailed) = false, want true")
	}
}

func TestAuthErrorIsErrAuthFailedAfterWrap(t *testing.T) {
	// 上层用 fmt.Errorf("...: %w", err) 包装后，链路仍须能命中 ErrAuthFailed。
	inner := AsAuth("token mismatch")
	wrapped := fmt.Errorf("auth failed: %w", inner)
	if !errors.Is(wrapped, ErrAuthFailed) {
		t.Fatalf("errors.Is(wrapped, ErrAuthFailed) = false, want true")
	}
}

func TestIsAuth(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"direct sentinel", ErrAuthFailed, true},
		{"AsAuth", AsAuth("token mismatch"), true},
		{"wrapped AsAuth", fmt.Errorf("auth failed: %w", AsAuth("token mismatch")), true},
		{"unrelated", errors.New("network timeout"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsAuth(c.err); got != c.want {
				t.Fatalf("IsAuth(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}
