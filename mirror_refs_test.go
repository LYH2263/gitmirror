package gitmirror

import (
	"context"
	"testing"

	"github.com/LYH2263/go-gitmirror/internal/refs"
)

// seedTips 在 dir 下用 refs.Store 写入 tips 并 flush，供 gitmirror.Open 读回。
func seedTips(t *testing.T, dir string, tips map[string]string) {
	t.Helper()
	s, err := refs.Open(dir)
	if err != nil {
		t.Fatalf("refs.Open: %v", err)
	}
	txn := s.Begin()
	for name, oid := range tips {
		if err := txn.Update(name, "", oid, true); err != nil {
			t.Fatalf("Update %s: %v", name, err)
		}
	}
	if err := txn.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// TestRefs_DoesNotAliasInternalBuf 锁定用户报告的 bug：
// 管理页拿到 Refs() 返回切片、把首条 Name 改成临时展示名后，
// 同进程再调一次 Refs() 不应返回被脏写的结果。
//
// 修复前 m.refsBuf 缓存命中时直接返回同一底层数组，
// 第二次 Refs() 的首条 Name 会被上次的临时名污染。
func TestRefs_DoesNotAliasInternalBuf(t *testing.T) {
	dir := t.TempDir()
	seedTips(t, dir, map[string]string{
		"refs/heads/main": "1111111111111111111111111111111111111111",
		"refs/heads/dev":  "2222222222222222222222222222222222222222",
	})

	m, err := Open(context.Background(), Options{Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })

	r1, err := m.Refs()
	if err != nil {
		t.Fatalf("Refs #1: %v", err)
	}
	if len(r1) != 2 {
		t.Fatalf("len(r1) = %d, want 2", len(r1))
	}
	// 排序后首条应为 dev。
	if got, want := r1[0].Name, "refs/heads/dev"; got != want {
		t.Fatalf("r1[0].Name = %q, want %q", got, want)
	}
	original := r1[0].Name
	tmp := "refs/heads/__tmp_display_name__"
	// 模拟前端高亮：改写返回切片首条 Name 与 NewOID。
	r1[0].Name = tmp
	r1[0].NewOID = ObjectID("deadbeef")

	// 同进程再调一次 Refs：不得返回被脏写的缓存。
	r2, err := m.Refs()
	if err != nil {
		t.Fatalf("Refs #2: %v", err)
	}
	if len(r2) != 2 {
		t.Fatalf("len(r2) = %d, want 2", len(r2))
	}
	if r2[0].Name == tmp {
		t.Fatalf("r2[0].Name = %q（被前次切片改动污染），want %q", r2[0].Name, original)
	}
	if r2[0].Name != original {
		t.Fatalf("r2[0].Name = %q, want %q", r2[0].Name, original)
	}
	if r2[0].NewOID == ObjectID("deadbeef") {
		t.Fatalf("r2[0].NewOID = %q（被前次切片改动污染）", r2[0].NewOID)
	}
	if r2[0].NewOID != ObjectID("2222222222222222222222222222222222222222") {
		t.Fatalf("r2[0].NewOID = %q, want 2222...", r2[0].NewOID)
	}

	// 两切片底层数组必须独立：对 r2 的改动不得回灌 r1。
	r2[1].Name = "refs/heads/__other_tmp__"
	if r1[1].Name == "refs/heads/__other_tmp__" {
		t.Fatalf("r1 与 r2 共享底层数组，Refs 未返回独立切片")
	}
}

// TestRefs_TwoCallsIndependent 验证连续两次 Refs 互不串味，
// 且在 tip 数量变化时不依赖旧缓存长度。
func TestRefs_TwoCallsIndependent(t *testing.T) {
	dir := t.TempDir()
	seedTips(t, dir, map[string]string{
		"refs/heads/main": "1111111111111111111111111111111111111111",
	})

	m, err := Open(context.Background(), Options{Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })

	a, err := m.Refs()
	if err != nil {
		t.Fatalf("Refs #1: %v", err)
	}
	b, err := m.Refs()
	if err != nil {
		t.Fatalf("Refs #2: %v", err)
	}
	if len(a) == 0 || len(b) == 0 {
		t.Fatalf("empty refs: a=%d b=%d", len(a), len(b))
	}
	a[0].Name = "MUTATED_A"
	if b[0].Name == "MUTATED_A" {
		t.Fatalf("b[0].Name 被前次切片改动污染 = %q", b[0].Name)
	}
	if b[0].Name != "refs/heads/main" {
		t.Fatalf("b[0].Name = %q, want refs/heads/main", b[0].Name)
	}
}
