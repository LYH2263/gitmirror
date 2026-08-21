package gitmirror

import (
	"github.com/LYH2263/go-gitmirror/internal/clone"
	"github.com/LYH2263/go-gitmirror/internal/validate"
)

// RemoteSpec 描述上游仓库与拉取参数。
type RemoteSpec struct {
	Name         string
	URL          string
	FetchRefs    []string // 空则默认 refs/heads/* + refs/tags/*
	ExtraHeaders []Header
	Auth         AuthSpec
	Depth        int // 0=完整；>0 浅克隆深度提示（Transport 可忽略）
	Prune        bool
}

// Clone 返回深拷贝，修改副本不影响原 Spec（含 ExtraHeaders / FetchRefs）。
func (r RemoteSpec) Clone() RemoteSpec {
	out := r
	out.FetchRefs = clone.Strings(r.FetchRefs)
	out.ExtraHeaders = cloneHeaders(r.ExtraHeaders)

	return out
}

// cloneHeaders 深拷贝 Header 切片；nil 保持 nil。
// ExtraHeaders 是 []Header 切片，值拷贝只复制切片头，底层数组共享，
// 调用方修改元素会写回原 Spec，故必须逐元素复制到新数组。
func cloneHeaders(src []Header) []Header {
	if src == nil {
		return nil
	}
	dst := make([]Header, len(src))
	copy(dst, src)
	return dst
}

// Validate 检查 URL 与 ref 模式。
func (r RemoteSpec) Validate() error {
	if err := validate.RemoteURL(r.URL); err != nil {
		return err
	}
	for _, pat := range r.FetchRefs {
		if err := validate.RefPattern(pat); err != nil {
			return err
		}
	}
	for _, h := range r.ExtraHeaders {
		if err := validate.HeaderKey(h.Key); err != nil {
			return err
		}
	}
	return nil
}

// HeaderMap 转为 map；同 key 后者覆盖。
func (r RemoteSpec) HeaderMap() map[string]string {
	if len(r.ExtraHeaders) == 0 {
		return nil
	}
	m := make(map[string]string, len(r.ExtraHeaders))
	for _, h := range r.ExtraHeaders {
		m[h.Key] = h.Value
	}
	return m
}
