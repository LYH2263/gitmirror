package gitmirror_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

func TestBug03_SyncAfterCloseNoPanic(t *testing.T) {
	url := "memory://bug03"
	root := t.TempDir()
	tr := fetch.NewMemory()
	tr.Seed(url, map[string]string{"refs/heads/main": "cccccccccccccccccccccccccccccccccccccccc"}, "")
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root:      root,
		Remote:    gitmirror.RemoteSpec{URL: url},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Sync after Close panicked via lockfile path: %v", rec)
		}
	}()
	_, err = m.Sync()
	if !errors.Is(err, gitmirror.ErrClosed) {
		t.Fatalf("want ErrClosed, got %v", err)
	}
}
