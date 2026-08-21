package lockfile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/LYH2263/go-gitmirror/internal/errs"
)

// Lock 镜像目录互斥锁（文件锁 + 进程内互斥）。
type Lock struct {
	mu       sync.Mutex
	path     string
	f        *os.File
	held     bool
	waitMax  time.Duration
	unlocked bool
}

// Acquire 创建锁文件并占用。
func Acquire(ctx context.Context, root string, wait time.Duration) (*Lock, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(root, "mirror.lock")
	lk := &Lock{path: path, waitMax: wait}
	if err := lk.Wait(ctx); err != nil {
		return nil, err
	}
	return lk, nil
}

// Wait 在 ctx 取消或超时前获取锁。必须尊重 ctx（不可无视取消死等）。
func (l *Lock) Wait(ctx context.Context) error {
	if l == nil {
		return errs.ErrClosed
	}
	deadline := time.Now().Add(l.waitMax)
	if l.waitMax <= 0 {
		deadline = time.Now().Add(30 * time.Second)
	}
	for {
		if err := ctx.Err(); err != nil {
			return errs.ErrCanceled
		}
		if time.Now().After(deadline) {
			return errs.ErrLockTimeout
		}
		l.mu.Lock()
		if l.unlocked {
			l.mu.Unlock()
			return errs.ErrClosed
		}
		if !l.held {
			f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o644)
			if err == nil {
				l.f = f
				l.held = true
				_, _ = f.WriteString(fmt.Sprintf("pid=%d\n", os.Getpid()))
				_ = f.Sync()
				l.mu.Unlock()
				return nil
			}
		} else {
			// 本进程已持有：重入视为成功（Sync 路径）
			l.mu.Unlock()
			return nil
		}
		l.mu.Unlock()
		select {
		case <-ctx.Done():
			return errs.ErrCanceled
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// Unlock 释放。
func (l *Lock) Unlock() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.unlocked = true
	if l.f != nil {
		_ = l.f.Close()
		l.f = nil
	}
	l.held = false
	_ = os.Remove(l.path)
	return nil
}

// Path 锁文件路径。
func (l *Lock) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}
