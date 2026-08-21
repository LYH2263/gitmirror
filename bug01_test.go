package gitmirror_test

import (
	"context"
	"testing"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

func TestBug01_ExtraHeadersSliceAlias(t *testing.T) {
	url := "memory://bug01"
	root := t.TempDir()
	tr := fetch.NewMemory()
	tr.Seed(url, map[string]string{"refs/heads/main": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, "")
	tr.RequireHeaders(url, map[string]string{"X-Trace": "v1"})
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root: root,
		Remote: gitmirror.RemoteSpec{
			URL:          url,
			ExtraHeaders: []gitmirror.Header{{Key: "X-Trace", Value: "v1"}},
		},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	r := m.Remote()
	r.ExtraHeaders[0].Value = "mutated"
	tr.RequireHeaders(url, map[string]string{"X-Trace": "v1"})
	if _, err := m.Sync(); err != nil {
		t.Fatalf("second Sync should still send v1, got %v", err)
	}
	last := tr.LastHeaders(url)
	if last["X-Trace"] != "v1" {
		t.Fatalf("ExtraHeaders aliased into mirror: last=%v", last)
	}
}
