package metrics

import (
	"sync"
	"time"
)

// Registry 简单计数器。
type Registry struct {
	mu sync.Mutex
	c  map[string]uint64
	lastSync time.Time
	lastErr  string
}

// Snapshot 对外快照。
type Snapshot struct {
	SyncBegin  uint64
	SyncOK     uint64
	PackBytes  uint64
	RefUpdates uint64
	LockOK     uint64
	LockFail   uint64
	LastSyncAt time.Time
	LastError  string
}

// New 创建。
func New() *Registry {
	return &Registry{c: make(map[string]uint64)}
}

// Inc 自增。
func (r *Registry) Inc(name string) {
	r.mu.Lock()
	r.c[name]++
	r.mu.Unlock()
}

// Add 累加。
func (r *Registry) Add(name string, n uint64) {
	r.mu.Lock()
	r.c[name] += n
	r.mu.Unlock()
}

// SetLastError 记录。
func (r *Registry) SetLastError(msg string) {
	r.mu.Lock()
	r.lastErr = msg
	r.mu.Unlock()
}

// SetLastSync 记录时间。
func (r *Registry) SetLastSync(t time.Time) {
	r.mu.Lock()
	r.lastSync = t
	r.mu.Unlock()
}

// Snapshot 拷贝。
func (r *Registry) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Snapshot{
		SyncBegin:  r.c["sync_begin"],
		SyncOK:     r.c["sync_ok"],
		PackBytes:  r.c["pack_bytes"],
		RefUpdates: r.c["ref_updates"],
		LockOK:     r.c["lock_ok"],
		LockFail:   r.c["lock_fail"],
		LastSyncAt: r.lastSync,
		LastError:  r.lastErr,
	}
}
