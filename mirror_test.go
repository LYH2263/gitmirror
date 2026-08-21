package gitmirror_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
	"github.com/LYH2263/go-gitmirror/internal/pack"
)

func oid(c byte) string {
	b := make([]byte, 40)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

func openTest(t *testing.T, url string, tips map[string]string) (*gitmirror.Mirror, *fetch.MemoryTransport) {
	t.Helper()
	root := t.TempDir()
	tr := fetch.NewMemory()
	tr.Seed(url, tips, "")
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root: root,
		Remote: gitmirror.RemoteSpec{
			Name: "origin",
			URL:  url,
		},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m, tr
}

func TestSyncUpdatesRefs(t *testing.T) {
	url := "memory://repo1"
	m, _ := openTest(t, url, map[string]string{
		"refs/heads/main": oid('a'),
	})
	rep, err := m.Sync()
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Updates) != 1 {
		t.Fatalf("updates=%d", len(rep.Updates))
	}
	got, err := m.Resolve("refs/heads/main")
	if err != nil || string(got) != oid('a') {
		t.Fatalf("resolve %v %v", got, err)
	}
}

func TestRefsReturnsIndependentSlice(t *testing.T) {
	url := "memory://repo2"
	m, _ := openTest(t, url, map[string]string{"refs/heads/main": oid('b')})
	if _, err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	a, err := m.Refs()
	if err != nil {
		t.Fatal(err)
	}
	a[0].Name = "tampered"
	b, err := m.Refs()
	if err != nil {
		t.Fatal(err)
	}
	if b[0].Name == "tampered" {
		t.Fatal("Refs shared backing array")
	}
}

func TestExtraHeadersSnapshotOnSync(t *testing.T) {
	url := "memory://repo3"
	m, tr := openTest(t, url, map[string]string{"refs/heads/main": oid('c')})
	tr.RequireHeaders(url, map[string]string{"X-Trace": "v1"})
	if err := m.SetRemote(gitmirror.RemoteSpec{
		Name: "origin",
		URL:  url,
		ExtraHeaders: []gitmirror.Header{{Key: "X-Trace", Value: "v1"}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	// 事后改 Remote 不应污染已完成的 Fetch；再次 Sync 用新头
	r := m.Remote()
	r.ExtraHeaders[0].Value = "v2"
	_ = m.SetRemote(r)
	tr.RequireHeaders(url, map[string]string{"X-Trace": "v2"})
	if _, err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	last := tr.LastHeaders(url)
	if last["X-Trace"] != "v2" {
		t.Fatalf("last=%v", last)
	}
}

func TestCloseThenSync(t *testing.T) {
	url := "memory://repo4"
	m, _ := openTest(t, url, map[string]string{"refs/heads/main": oid('d')})
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := m.Sync()
	if !errors.Is(err, gitmirror.ErrClosed) {
		t.Fatalf("got %v", err)
	}
}

func TestNoTransport(t *testing.T) {
	root := t.TempDir()
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root:   root,
		Remote: gitmirror.RemoteSpec{URL: "memory://x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	_, err = m.Sync()
	if !errors.Is(err, gitmirror.ErrNoTransport) {
		t.Fatalf("got %v", err)
	}
}

func TestAuthFailedWrapped(t *testing.T) {
	url := "memory://secure"
	root := t.TempDir()
	tr := fetch.NewMemory()
	tr.Seed(url, map[string]string{"refs/heads/main": oid('e')}, "secret")
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root:      root,
		Remote:    gitmirror.RemoteSpec{URL: url, Auth: gitmirror.AuthSpec{Token: "wrong"}},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	_, err = m.Sync()
	if !errors.Is(err, gitmirror.ErrAuthFailed) {
		t.Fatalf("got %v", err)
	}
}

func TestSyncContextCanceled(t *testing.T) {
	url := "memory://ctx"
	m, _ := openTest(t, url, map[string]string{"refs/heads/main": oid('f')})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := m.SyncContext(ctx)
	if !errors.Is(err, gitmirror.ErrCanceled) {
		t.Fatalf("got %v", err)
	}
}

func TestPackVerify(t *testing.T) {
	good := pack.BuildMinimal()
	if err := pack.Verify(good); err != nil {
		t.Fatal(err)
	}
	bad := append([]byte(nil), good...)
	bad[len(bad)-1] ^= 0xff
	if err := pack.Verify(bad); err == nil {
		t.Fatal("expected corrupt")
	}
	n, err := pack.ReceiveAndVerify(context.Background(), t.TempDir(), good, 0)
	if err != nil || n == 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestPrune(t *testing.T) {
	url := "memory://prune"
	m, tr := openTest(t, url, map[string]string{
		"refs/heads/main": oid('1'),
		"refs/heads/old":  oid('2'),
	})
	if _, err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	tr.Seed(url, map[string]string{"refs/heads/main": oid('1')}, "")
	if err := m.SetRemote(gitmirror.RemoteSpec{URL: url, Prune: true}); err != nil {
		t.Fatal(err)
	}
	rep, err := m.Sync()
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.PrunedRefs) == 0 {
		t.Fatal("expected prune")
	}
	_ = filepath.Separator
	_ = time.Second
}
