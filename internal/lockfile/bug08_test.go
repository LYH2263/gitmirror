package lockfile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LYH2263/go-gitmirror/internal/errs"
	"github.com/LYH2263/go-gitmirror/internal/lockfile"
)

func TestBug08_LockWaitHonorsContext(t *testing.T) {
	// 已取消的 ctx 进入 Acquire→Wait；须立刻 ErrCanceled（Sync 等锁同路径传 ctx）。
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	_, err := lockfile.Acquire(ctx, t.TempDir(), 5*time.Second)
	if !errors.Is(err, errs.ErrCanceled) {
		t.Fatalf("want ErrCanceled, got %v", err)
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("Wait ignored ctx and blocked toward LockTimeout")
	}
}
