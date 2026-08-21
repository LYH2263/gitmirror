package gitmirror

import "time"

// ObjectID 为 40 位十六进制 SHA-1（或测试用伪 OID）。
type ObjectID string

// ZeroOID 表示删除或尚无对象。
const ZeroOID ObjectID = "0000000000000000000000000000000000000000"

// RefUpdate 描述一次本地 tip 变更。
type RefUpdate struct {
	Name   string
	OldOID ObjectID
	NewOID ObjectID
	Force  bool
}

// SyncReport 是一次 Sync/SyncContext 的结果摘要。
type SyncReport struct {
	RemoteURL   string
	StartedAt   time.Time
	FinishedAt  time.Time
	Updates     []RefUpdate
	PackBytes   int64
	PackCount   int
	PrunedRefs  []string
	FetchedTips map[string]ObjectID
}

// MirrorStats 运行时计数。
type MirrorStats struct {
	Syncs       uint64
	Fetches     uint64
	PackBytes   uint64
	RefUpdates  uint64
	LockWaits   uint64
	LastSyncAt  time.Time
	LastError   string
	RefCount    int
}

// AuthSpec 远端认证；Token 优先于 User/Pass。
type AuthSpec struct {
	Token string
	User  string
	Pass  string
}

// Header 额外 HTTP 头（fetch 协议协商用）。
type Header struct {
	Key   string
	Value string
}
