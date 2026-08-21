package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// newStoreWithTips 构造一个带若干 tip 的 Store，用于隔离性测试。
func newStoreWithTips(t *testing.T, tips map[string]string) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	txn := s.Begin()
	for name, oid := range tips {
		if err := txn.Update(name, "", oid, true); err != nil {
			t.Fatalf("Update %s: %v", name, err)
		}
	}
	if err := txn.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return s
}

// TestList_DoesNotAliasInternalBuffer 锁定 List 的契约：
// 返回切片必须每次独立分配，修改它不得污染后续 List 的结果，
// 也不得与内部状态串味。
func TestList_DoesNotAliasInternalBuffer(t *testing.T) {
	s := newStoreWithTips(t, map[string]string{
		"refs/heads/main": "1111111111111111111111111111111111111111",
		"refs/heads/dev":  "2222222222222222222222222222222222222222",
	})

	// 第一次取切片，模拟前端把首条 Name 改成临时展示名。
	first := s.List()
	if len(first) != 2 {
		t.Fatalf("len(first) = %d, want 2", len(first))
	}
	if first[0].Name != "refs/heads/dev" {
		t.Fatalf("first[0].Name = %q, want refs/heads/dev", first[0].Name)
	}
	original := first[0].Name
	mutated := "refs/heads/__tmp_display_name__"
	first[0].Name = mutated
	first[0].OID = "deadbeef"

	// 第二次取切片：内部表不应被第一次的修改污染。
	second := s.List()
	if len(second) != 2 {
		t.Fatalf("len(second) = %d, want 2", len(second))
	}
	if second[0].Name == mutated {
		t.Fatalf("second[0].Name = %q (被脏写)，want %q", second[0].Name, original)
	}
	if second[0].Name != original {
		t.Fatalf("second[0].Name = %q, want %q", second[0].Name, original)
	}
	if second[0].OID == "deadbeef" {
		t.Fatalf("second[0].OID = %q (被脏写)", second[0].OID)
	}
	if second[0].OID != "2222222222222222222222222222222222222222" {
		t.Fatalf("second[0].OID = %q, want 2222...", second[0].OID)
	}

	// 两切片底层数组必须独立：first 的改动不应经由 second 可见，
	// second 的改动也不应回灌 first。
	second[1].Name = "refs/heads/__other_tmp__"
	if first[1].Name == "refs/heads/__other_tmp__" {
		t.Fatalf("first 与 second 共享底层数组，List 未返回独立切片")
	}
}

// TestList_TwoCallsAreIndependent 验证连续两次 List 互不串味
// （回归 s.listBuf 复用导致第二次覆盖第一次内容的缺陷）。
func TestList_TwoCallsAreIndependent(t *testing.T) {
	s := newStoreWithTips(t, map[string]string{
		"refs/heads/main": "1111111111111111111111111111111111111111",
	})
	a := s.List()
	b := s.List()
	a[0].Name = "MUTATED_A"
	if b[0].Name == "MUTATED_A" {
		t.Fatalf("b[0].Name 被前次切片改动污染 = %q", b[0].Name)
	}
	if b[0].Name != "refs/heads/main" {
		t.Fatalf("b[0].Name = %q, want refs/heads/main", b[0].Name)
	}
}

// TestList_GrowsBeyondPreviousCapacity 验证当 tip 数量超过上次 List
// 的容量时（曾经 listBuf[:0] 复用旧底层数组的场景），结果仍正确。
func TestList_GrowsBeyondPreviousCapacity(t *testing.T) {
	s := newStoreWithTips(t, map[string]string{
		"refs/heads/a": "a0000000000000000000000000000000000000000",
	})
	_ = s.List()

	// 追加大量 tip 后再 List，确保不依赖旧缓冲长度。
	for i := 0; i < 64; i++ {
		name := fmt.Sprintf("refs/heads/b%02d", i)
		txn := s.Begin()
		if err := txn.Update(name, "", "b0000000000000000000000000000000000000000", true); err != nil {
			t.Fatalf("Update %s: %v", name, err)
		}
		if err := txn.Commit(); err != nil {
			t.Fatalf("Commit: %v", err)
		}
	}
	got := s.List()
	if len(got) != 65 {
		t.Fatalf("len = %d, want 65", len(got))
	}
	// 验证排序且内容未被截断/串味。
	for i := 1; i < len(got); i++ {
		if got[i-1].Name >= got[i].Name {
			t.Fatalf("未排序或重复: %q >= %q at %d", got[i-1].Name, got[i].Name, i)
		}
	}
	// 确认首条仍是 a（磁盘上的原始 tip 未被复用缓冲覆盖）。
	if _, err := os.Stat(filepath.Join(s.root, filepath.FromSlash("refs/heads/a"))); err != nil {
		t.Fatalf("原始 ref 文件丢失: %v", err)
	}
}
