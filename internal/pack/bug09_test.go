package pack_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LYH2263/go-gitmirror/internal/pack"
)

func TestBug09_PackTempCloseBeforeRename(t *testing.T) {
	root := t.TempDir()
	data := pack.BuildMinimal()
	n, err := pack.ReceiveAndVerify(context.Background(), root, data, 0)
	if err != nil {
		t.Fatalf("ReceiveAndVerify failed (temp file must Close before Rename): %v", err)
	}
	if n == 0 {
		t.Fatal("wrote 0 bytes")
	}
	entries, err := os.ReadDir(filepath.Join(root, "objects", "pack"))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".pack" {
			found = true
		}
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("leftover tmp: %s", e.Name())
		}
	}
	if !found {
		t.Fatal("final pack missing")
	}
}
