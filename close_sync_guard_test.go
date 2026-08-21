package gitmirror

import (
	"context"
	"errors"
	"testing"

	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

// TestSyncContextAfterCloseReturnsErrClosed 是 bug3 的回归测试：
// 镜像目录维护完 Close() 后紧接着自动化触发一次 Sync()，应稳定返回
// ErrClosed，而非在 m.lock.Wait 上空指针 panic。
//
// 根因：SyncContext 在 m.mu.Lock() 之后未调用 guard() 即直接
// m.lock.Wait(ctx)；而 Close() 已将 m.lock 置 nil（close.go:33），
// 因此 nil 解引用 panic。修复为在加锁后立即 guard() 拦截。
func TestSyncContextAfterCloseReturnsErrClosed(t *testing.T) {
	root := t.TempDir()
	tr := fetch.NewMemory()
	tr.Seed("memory://up", map[string]string{
		"refs/heads/main": "0000000000000000000000000000000000000001",
	}, "")

	m, err := Open(context.Background(), Options{
		Root:      root,
		Remote:    RemoteSpec{URL: "memory://up"},
		Transport: tr,
		LockWait:  0,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	// 先做一次正常 Sync 确保路径通畅。
	if _, err := m.Sync(); err != nil {
		t.Fatalf("initial Sync: %v", err)
	}

	// 关闭镜像：此后 m.lock == nil、m.closed == true。
	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// 维护完点关闭后，自动化紧接着触发一次 Sync —— 这里即原先 panic 点。
	// 用 recover 兜底：修复前会 panic，修复后应干净返回 ErrClosed。
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Sync after Close panicked (expected ErrClosed): %v", r)
			}
		}()
		_, err := m.Sync()
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed after Close, got %v", err)
		}
	}()

	// SyncContext 同路径，同样校验。
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("SyncContext after Close panicked (expected ErrClosed): %v", r)
			}
		}()
		_, err := m.SyncContext(context.Background())
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed after Close, got %v", err)
		}
	}()
}

// Prune 同样在 guard() 之后才碰 m.lock.Wait，关闭后也应稳返 ErrClosed。
func TestPruneAfterCloseReturnsErrClosed(t *testing.T) {
	root := t.TempDir()
	tr := fetch.NewMemory()
	tr.Seed("memory://up", map[string]string{
		"refs/heads/main": "0000000000000000000000000000000000000001",
	}, "")

	m, err := Open(context.Background(), Options{
		Root:      root,
		Remote:    RemoteSpec{URL: "memory://up"},
		Transport: tr,
		LockWait:  0,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := m.Sync(); err != nil {
		t.Fatalf("initial Sync: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Prune after Close panicked (expected ErrClosed): %v", r)
			}
		}()
		_, err := m.Prune(nil)
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed after Close, got %v", err)
		}
	}()
}
