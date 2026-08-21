package admin

import (
	"time"

	"github.com/LYH2263/go-gitmirror"
)

// Dashboard 管理页聚合视图。
type Dashboard struct {
	RemoteURL  string            `json:"remote_url"`
	RefCount   int               `json:"ref_count"`
	Stats      gitmirror.MirrorStats `json:"stats"`
	Refs       []gitmirror.RefUpdate `json:"refs"`
	GeneratedAt time.Time        `json:"generated_at"`
}

// FromMirror 构建。
func FromMirror(m *gitmirror.Mirror) (*Dashboard, error) {
	refs, err := m.Refs()
	if err != nil {
		return nil, err
	}
	r := m.Remote()
	st := m.Stats()
	return &Dashboard{
		RemoteURL:   r.URL,
		RefCount:    len(refs),
		Stats:       st,
		Refs:        refs,
		GeneratedAt: time.Now().UTC(),
	}, nil
}
