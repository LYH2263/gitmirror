package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// SyncState 最近一次同步元数据。
type SyncState struct {
	LastRemote    string    `json:"last_remote"`
	LastSyncAt    time.Time `json:"last_sync_at"`
	LastPackBytes int64     `json:"last_pack_bytes"`
	SyncCount     uint64    `json:"sync_count"`
}

// LoadSyncState 读取；不存在则空状态。
func LoadSyncState(root string) (*SyncState, error) {
	p := filepath.Join(root, "sync-state.json")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &SyncState{}, nil
		}
		return nil, err
	}
	var st SyncState
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// SaveSyncState 原子写。
func SaveSyncState(root string, st *SyncState) error {
	if st == nil {
		return nil
	}
	st.SyncCount++
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(root, "sync-state.json")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
