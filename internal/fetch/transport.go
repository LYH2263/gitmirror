package fetch

import (
	"context"

	"github.com/LYH2263/go-gitmirror/internal/clone"
)

// Auth 传输层认证。
type Auth struct {
	Token string
	User  string
	Pass  string
}

// Request 一次 fetch 协商请求。
type Request struct {
	URL          string
	WantPatterns []string
	Have         map[string]string
	Headers      map[string]string
	Auth         Auth
	Depth        int
}

// Clone 深拷贝请求（Headers/Have/WantPatterns 隔离）。
func (r Request) Clone() Request {
	out := r
	out.WantPatterns = clone.Strings(r.WantPatterns)
	out.Have = clone.StringMap(r.Have)
	out.Headers = clone.StringMap(r.Headers)
	return out
}

// Result 远端 tip 与 pack 载荷。
type Result struct {
	Tips  map[string]string
	Packs [][]byte
}

// Transport 抽象协商与下载；生产可用 HTTP smart protocol，测试用 Memory。
type Transport interface {
	Fetch(ctx context.Context, req Request) (*Result, error)
}
