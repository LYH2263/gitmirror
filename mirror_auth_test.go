package gitmirror

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

// newMirrorForTest 打开一个临时镜像目录，并注入 MemoryTransport。
func newMirrorForTest(t *testing.T, remote RemoteSpec, tr *fetch.MemoryTransport) *Mirror {
	t.Helper()
	dir := t.TempDir()
	// Open 需要 RemoteSpec.URL 通过校验，故传入一个合法远端。
	m, err := Open(context.Background(), Options{Root: dir, Remote: remote})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })
	if err := m.SetTransport(tr); err != nil {
		t.Fatalf("SetTransport: %v", err)
	}
	return m
}

func TestSync_AuthFailedIsErrAuthFailed(t *testing.T) {
	const url = "https://example.invalid/upstream.git"
	tr := fetch.NewMemory()
	// 注册一个需要 token 的远端；Sync 故意传错 token 触发认证失败。
	tr.Seed(url, map[string]string{"refs/heads/main": "0123456789abcdef0123456789abcdef01234567"}, "correct-token")

	m := newMirrorForTest(t, RemoteSpec{
		URL:  url,
		Auth: AuthSpec{Token: "wrong-token"},
	}, tr)

	_, err := m.Sync()
	if err == nil {
		t.Fatal("Sync: expected auth error, got nil")
	}
	// 业务侧告警分支依赖此判断。修复前为 false。
	if !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("errors.Is(err, ErrAuthFailed) = false, want true (err=%v)", err)
	}
}

func TestSync_AuthFailedIsErrAuthFailed_HeaderMismatch(t *testing.T) {
	const url = "https://example.invalid/headers.git"
	tr := fetch.NewMemory()
	// 注意：Seed 会整体覆盖 memRemote，因此先 Seed 再 RequireHeaders。
	tr.Seed(url, map[string]string{"refs/heads/main": "0123456789abcdef0123456789abcdef01234567"}, "")
	tr.RequireHeaders(url, map[string]string{"X-Signature": "abc"})

	m := newMirrorForTest(t, RemoteSpec{
		URL:          url,
		ExtraHeaders: []Header{{Key: "X-Signature", Value: "wrong"}},
	}, tr)

	_, err := m.Sync()
	if err == nil {
		t.Fatal("Sync: expected auth error, got nil")
	}
	if !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("errors.Is(err, ErrAuthFailed) = false, want true (err=%v)", err)
	}
}

func TestSync_OK_NoAuthError(t *testing.T) {
	const url = "https://example.invalid/ok.git"
	tr := fetch.NewMemory()
	tr.Seed(url, map[string]string{"refs/heads/main": "0123456789abcdef0123456789abcdef01234567"}, "")

	m := newMirrorForTest(t, RemoteSpec{URL: url}, tr)

	rep, err := m.Sync()
	if err != nil {
		t.Fatalf("Sync: unexpected error: %v", err)
	}
	if rep == nil {
		t.Fatal("Sync: nil report")
	}
	if errors.Is(err, ErrAuthFailed) {
		t.Fatal("successful Sync must not be flagged as auth failed")
	}
}

// 防止 lint 报未使用 import（部分场景下 time 未直接引用）。
var _ = time.Second
var _ = os.PathSeparator
var _ = filepath.Separator
