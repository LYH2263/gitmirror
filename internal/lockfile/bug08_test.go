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
	root := t.TempDir()
	holder, err := lockfile.Acquire(context.Background(), root, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := lockfile.Acquire(ctx, root, 5*time.Second)
		done <- err
	}()
	time.Sleep(30 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, errs.ErrCanceled) {
			t.Fatalf("want ErrCanceled, got %v", err)
		}
	case <-time.After(800 * time.Millisecond):
		t.Fatal("Wait ignored ctx cancel (Sync 等锁同路径)")
	}
}
