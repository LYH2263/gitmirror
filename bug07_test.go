package gitmirror_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/errs"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

type gateTransport struct {
	inner   *fetch.MemoryTransport
	started chan struct{}
	gate    chan struct{}
}

func (g *gateTransport) Fetch(ctx context.Context, req fetch.Request) (*fetch.Result, error) {
	close(g.started)
	select {
	case <-g.gate:
	case <-ctx.Done():
		return nil, errs.ErrCanceled
	}
	return g.inner.Fetch(ctx, req)
}

func TestBug07_SyncHonorsContextCancel(t *testing.T) {
	url := "memory://bug07"
	root := t.TempDir()
	inner := fetch.NewMemory()
	inner.Seed(url, map[string]string{"refs/heads/main": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"}, "")
	started := make(chan struct{})
	gate := make(chan struct{})
	tr := &gateTransport{inner: inner, started: started, gate: gate}
	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root:      root,
		Remote:    gitmirror.RemoteSpec{URL: url},
		Transport: tr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := m.SyncContext(ctx)
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("Fetch did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, gitmirror.ErrCanceled) {
			t.Fatalf("want ErrCanceled, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		close(gate)
		t.Fatal("SyncContext ignored cancel and kept downloading")
	}
	close(gate)
}
