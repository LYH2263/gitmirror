package lockfile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LYH2263/go-gitmirror/internal/errs"
	"github.com/LYH2263/go-gitmirror/internal/lockfile"
)

func TestWaitRespectsCancel(t *testing.T) {
	root := t.TempDir()
	lk, err := lockfile.Acquire(context.Background(), root, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_ = lk.Unlock()
	lk2, err := lockfile.Acquire(context.Background(), root, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer lk2.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// 新 Lock 在已 unlocked 路径
	lk3 := lk // unlocked
	err = lk3.Wait(ctx)
	if !errors.Is(err, errs.ErrCanceled) && !errors.Is(err, errs.ErrClosed) {
		// Acquire 新锁时 cancel
		_, err = lockfile.Acquire(ctx, t.TempDir(), time.Millisecond)
		if !errors.Is(err, errs.ErrCanceled) {
			t.Fatalf("got %v", err)
		}
	}
}
