package gitmirror_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

func TestBug05_AuthFailedWraps(t *testing.T) {
	url := "memory://bug05"
	root := t.TempDir()
	tr := fetch.NewMemory()
	tr.Seed(url, map[string]string{"refs/heads/main": "dddddddddddddddddddddddddddddddddddddddd"}, "secret-token")
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
	if err == nil {
		t.Fatal("expected auth error")
	}
	if !errors.Is(err, gitmirror.ErrAuthFailed) {
		t.Fatalf("want errors.Is ErrAuthFailed, got %v", err)
	}
}
