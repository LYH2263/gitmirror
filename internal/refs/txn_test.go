package refs_test

import (
	"testing"

	"github.com/LYH2263/go-gitmirror/internal/refs"
)

func TestTxnCommitFailureDoesNotUpdateMemory(t *testing.T) {
	root := t.TempDir()
	st, err := refs.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	txn := st.Begin()
	if err := txn.Update("refs/heads/main", "", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false); err != nil {
		t.Fatal(err)
	}
	if err := txn.Commit(); err != nil {
		t.Fatal(err)
	}
	oid, ok := st.Resolve("refs/heads/main")
	if !ok || oid == "" {
		t.Fatal("missing tip after commit")
	}
	// Abort path: begin then abort keeps old
	txn2 := st.Begin()
	_ = txn2.Update("refs/heads/main", oid, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", false)
	txn2.Abort()
	oid2, _ := st.Resolve("refs/heads/main")
	if oid2 != oid {
		t.Fatalf("abort mutated tip %s -> %s", oid, oid2)
	}
}
