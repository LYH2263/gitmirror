package gitmirror

import (
	"context"
	"testing"
	"time"

	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

// TestCloneIsolatesExtraHeaders 验证 RemoteSpec.Clone 对 ExtraHeaders 做深拷贝。
// 修改副本的元素不得写回原 Spec（切片底层数组必须独立）。
func TestCloneIsolatesExtraHeaders(t *testing.T) {
	src := RemoteSpec{
		Name: "origin",
		URL:  "memory://x",
		ExtraHeaders: []Header{
			{Key: "X-Trace", Value: "orig"},
			{Key: "Authorization", Value: "Bearer s3cr3t"},
		},
	}

	cp := src.Clone()
	// 改副本元素值
	cp.ExtraHeaders[0].Value = "TAMPERED"
	cp.ExtraHeaders[1] = Header{Key: "X-Evil", Value: "evil"}

	if got := src.ExtraHeaders[0].Value; got != "orig" {
		t.Fatalf("原 Spec 的 ExtraHeaders[0] 被副本污染: got %q want %q", got, "orig")
	}
	if got := src.ExtraHeaders[1].Key; got != "Authorization" {
		t.Fatalf("原 Spec 的 ExtraHeaders[1] 被副本污染: got %q want %q", got, "Authorization")
	}
}

// TestRemoteEditDoesNotLeakIntoSync 复现报告场景：
// 配好 ExtraHeaders → Sync → 经 Remote() 取出改 X-Trace 打调试日志（不再 SetRemote）
// → 再 Sync。上游第二次收到的头必须是原值，鉴权/协商不破。
func TestRemoteEditDoesNotLeakIntoSync(t *testing.T) {
	const url = "memory://headers-leak"
	tr := fetch.NewMemory()
	tr.Seed(url, map[string]string{"refs/heads/main": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, "")
	// 上游要求 X-Trace == trace-v1
	tr.RequireHeaders(url, map[string]string{"X-Trace": "trace-v1"})

	dir := t.TempDir()
	m, err := Open(context.Background(), Options{
		Root: dir,
		Remote: RemoteSpec{
			Name: "origin",
			URL:  url,
			ExtraHeaders: []Header{
				{Key: "X-Trace", Value: "trace-v1"},
			},
		},
		Transport: tr,
		LockWait:  5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer m.Close()

	// 第一次 Sync：应成功，X-Trace=trace-v1 上游可见。
	if _, err := m.SyncContext(context.Background()); err != nil {
		t.Fatalf("首次 Sync 失败: %v", err)
	}

	// 取出 Spec 改 X-Trace 打调试日志，但不 SetRemote 回去。
	spec := m.Remote()
	idx := -1
	for i, h := range spec.ExtraHeaders {
		if h.Key == "X-Trace" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("Remote() 未返回 X-Trace 头")
	}
	spec.ExtraHeaders[idx].Value = "debug-override"

	// 第二次 Sync：上游仍应收到 X-Trace=trace-v1。
	if _, err := m.SyncContext(context.Background()); err != nil {
		t.Fatalf("二次 Sync 因头被污染失败（鉴权/协商对不上）: %v", err)
	}

	last := tr.LastHeaders(url)
	if got := last["X-Trace"]; got != "trace-v1" {
		t.Fatalf("上游第二次收到被改过的头: X-Trace=%q want %q", got, "trace-v1")
	}
}
