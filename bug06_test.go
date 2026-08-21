package gitmirror_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

func TestBug06_TxnCommitRollback(t *testing.T) {
	url := "memory://bug06"
	root := t.TempDir()
	tr := fetch.NewMemory()
	oid1 := "1111111111111111111111111111111111111111"
	oid2 := "2222222222222222222222222222222222222222"
	tr.Seed(url, map[string]string{"refs/heads/main": oid1}, "")
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
	got, err := m.Resolve("refs/heads/main")
	if err != nil || string(got) != oid1 {
		t.Fatalf("first tip %v %v", got, err)
	}
	// 让后续 commit 写 ref 文件失败：把 refs/heads/main 变成目录
	refPath := filepath.Join(root, "refs", "heads", "main")
	_ = os.Remove(refPath)
	if err := os.Mkdir(refPath, 0o755); err != nil {
		t.Fatal(err)
	}
	tr.Seed(url, map[string]string{"refs/heads/main": oid2}, "")
	_, err = m.Sync()
	if err == nil {
		t.Fatal("expected persist error")
	}
	got2, err := m.Resolve("refs/heads/main")
	if err != nil {
		t.Fatal(err)
	}
	if string(got2) == oid2 {
		t.Fatal("memory tip updated despite commit write failure")
	}
	if string(got2) != oid1 {
		t.Fatalf("want old tip %s got %s", oid1, got2)
	}
}
