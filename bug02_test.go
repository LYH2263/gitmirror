package gitmirror_test

import (
	"context"
	"testing"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

func TestBug02_RefsSliceAlias(t *testing.T) {
	url := "memory://bug02"
	root := t.TempDir()
	tr := fetch.NewMemory()
	tr.Seed(url, map[string]string{"refs/heads/main": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}, "")
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root:      root,
		Remote:    gitmirror.RemoteSpec{URL: url},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	a, err := m.Refs()
	if err != nil || len(a) == 0 {
		t.Fatalf("refs=%v err=%v", a, err)
	}
	a[0].Name = "tampered"
	b, err := m.Refs()
	if err != nil {
		t.Fatal(err)
	}
	if b[0].Name == "tampered" {
		t.Fatal("Refs returned shared slice with prior caller mutation")
	}
}
