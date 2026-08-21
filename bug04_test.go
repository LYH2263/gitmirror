package gitmirror_test

import (
	"context"
	"testing"
	"time"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

func TestBug04_NilExtraHeadersSyncNoHalfRefs(t *testing.T) {
	url := "memory://bug04-headers"
	root := t.TempDir()
	tr := fetch.NewMemory()
	oid := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tr.Seed(url, map[string]string{"refs/heads/main": oid}, "")
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root: root,
		Remote: gitmirror.RemoteSpec{
			URL:          url,
			ExtraHeaders: nil,
		},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		done := make(chan struct{})
		go func() {
			_ = m.Close()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			// plant panic 路径可能未释放锁；勿让 Close 拖死整个 go test 进程
		}
	}()
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Sync panicked on nil ExtraHeaders map write: %v", rec)
		}
	}()
	if _, err := m.Sync(); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	got, err := m.Resolve("refs/heads/main")
	if err != nil || string(got) != oid {
		t.Fatalf("refs half-updated or missing: %v %v", got, err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := m.Sync()
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("second Sync after headers path: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("second Sync hung — refs txn/lock likely not released after panic path")
	}
}
