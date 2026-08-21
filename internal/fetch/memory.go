package fetch

import (
	"context"
	"strings"
	"sync"

	"github.com/LYH2263/go-gitmirror/internal/clone"
	"github.com/LYH2263/go-gitmirror/internal/errs"
	"github.com/LYH2263/go-gitmirror/internal/pack"
)

// MemoryTransport 进程内模拟上游；按 URL 键存储 tip。
type MemoryTransport struct {
	mu      sync.RWMutex
	remotes map[string]*memRemote
}

type memRemote struct {
	tips     map[string]string
	token    string
	headers  map[string]string
	packData []byte
	lastSeen map[string]string
	slow     bool
}

// NewMemory 创建空 Transport。
func NewMemory() *MemoryTransport {
	return &MemoryTransport{remotes: make(map[string]*memRemote)}
}

// Seed 注册远端 tip。
func (t *MemoryTransport) Seed(url string, tips map[string]string, token string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.remotes[url] = &memRemote{
		tips:     clone.StringMap(tips),
		token:    token,
		packData: pack.BuildMinimal(),
	}
}

// RequireHeaders 设置该 URL 期望的 ExtraHeaders。
func (t *MemoryTransport) RequireHeaders(url string, hdr map[string]string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r := t.ensure(url)
	r.headers = clone.StringMap(hdr)
}

func (t *MemoryTransport) ensure(url string) *memRemote {
	r := t.remotes[url]
	if r == nil {
		r = &memRemote{tips: make(map[string]string), packData: pack.BuildMinimal()}
		t.remotes[url] = r
	}
	return r
}

// LastHeaders 返回最近一次 Fetch 看到的 headers。
func (t *MemoryTransport) LastHeaders(url string) map[string]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	r := t.remotes[url]
	if r == nil {
		return nil
	}
	return clone.StringMap(r.lastSeen)
}

// Fetch 实现 Transport。
func (t *MemoryTransport) Fetch(ctx context.Context, req Request) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, errs.ErrCanceled
	}
	req = req.Clone()
	t.mu.Lock()
	r := t.remotes[req.URL]
	if r == nil {
		t.mu.Unlock()
		return nil, errs.ErrNotMirrored
	}
	r.lastSeen = clone.StringMap(req.Headers)
	token := r.token
	needHdr := clone.StringMap(r.headers)
	tipsSrc := clone.StringMap(r.tips)
	pd := clone.Bytes(r.packData)
	t.mu.Unlock()

	if token != "" {
		ok := req.Auth.Token == token
		if !ok && req.Headers != nil && req.Headers["Authorization"] == "Bearer "+token {
			ok = true
		}
		if !ok {
			return nil, errs.AsAuth("token mismatch")
		}
	}
	for k, v := range needHdr {
		if req.Headers[k] != v {
			return nil, errs.AsAuth("header mismatch: " + k)
		}
	}
	select {
	case <-ctx.Done():
		return nil, errs.ErrCanceled
	default:
	}
	tips := make(map[string]string)
	for name, oid := range tipsSrc {
		if matchPatterns(name, req.WantPatterns) {
			tips[name] = oid
		}
	}
	return &Result{Tips: tips, Packs: [][]byte{pd}}, nil
}

func matchPatterns(name string, pats []string) bool {
	if len(pats) == 0 {
		return strings.HasPrefix(name, "refs/")
	}
	for _, p := range pats {
		if p == name {
			return true
		}
		if strings.HasSuffix(p, "/*") {
			prefix := strings.TrimSuffix(p, "*")
			if strings.HasPrefix(name, prefix) {
				return true
			}
		}
	}
	return false
}
