package gitmirror_test

import (
	"context"
	"testing"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

func TestBug10_CloseFlushBeforeUnlock(t *testing.T) {
	url := "memory://bug10"
	root := t.TempDir()
	tr := fetch.NewMemory()
	oid := "ffffffffffffffffffffffffffffffffffffffff"
	tr.Seed(url, map[string]string{"refs/heads/main": oid}, "")
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root:      root,
		Remote:    gitmirror.RemoteSpec{URL: url},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	m2, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root:      root,
		Remote:    gitmirror.RemoteSpec{URL: url},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer m2.Close()
	got, err := m2.Resolve("refs/heads/main")
	if err != nil {
		t.Fatalf("reopen lost tip after Close without flush: %v", err)
	}
	if string(got) != oid {
		t.Fatalf("want %s got %s", oid, got)
	}
}
