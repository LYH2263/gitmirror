package gitmirror

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/LYH2263/go-gitmirror/internal/audit"
	"github.com/LYH2263/go-gitmirror/internal/clone"
	"github.com/LYH2263/go-gitmirror/internal/errs"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
	"github.com/LYH2263/go-gitmirror/internal/lockfile"
	"github.com/LYH2263/go-gitmirror/internal/metrics"
	"github.com/LYH2263/go-gitmirror/internal/pack"
	"github.com/LYH2263/go-gitmirror/internal/persist"
	"github.com/LYH2263/go-gitmirror/internal/refs"
	"github.com/LYH2263/go-gitmirror/internal/validate"
)

// Mirror 本地裸仓镜像门面；线程安全。
type Mirror struct {
	mu        sync.RWMutex
	closed    bool
	root      string
	remote    RemoteSpec
	transport fetch.Transport
	refs      *refs.Store
	lock      *lockfile.Lock
	audit     *audit.Log
	met       *metrics.Registry
	lockWait  time.Duration
	state     *persist.SyncState
}

// Open 打开或创建镜像目录。
func Open(ctx context.Context, opt Options) (*Mirror, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if opt.Root == "" {
		return nil, fmt.Errorf("%w: empty root", errs.ErrInvalidRemote)
	}
	remote := opt.Remote.Clone()
	if remote.URL != "" {
		if err := remote.Validate(); err != nil {
			return nil, err
		}
	}
	wait := opt.LockWait
	if wait <= 0 {
		wait = 30 * time.Second
	}
	lk, err := lockfile.Acquire(ctx, opt.Root, wait)
	if err != nil {
		return nil, err
	}
	st, err := refs.Open(opt.Root)
	if err != nil {
		_ = lk.Unlock()
		return nil, err
	}
	al, err := audit.Open(opt.Root)
	if err != nil {
		_ = st.Close()
		_ = lk.Unlock()
		return nil, err
	}
	state, err := persist.LoadSyncState(opt.Root)
	if err != nil {
		_ = al.Close()
		_ = st.Close()
		_ = lk.Unlock()
		return nil, err
	}
	m := &Mirror{
		root:      opt.Root,
		remote:    remote,
		transport: opt.Transport,
		refs:      st,
		lock:      lk,
		audit:     al,
		met:       metrics.New(),
		lockWait:  wait,
		state:     state,
	}
	return m, nil
}

// SetRemote 替换远端配置（深拷贝）。
func (m *Mirror) SetRemote(r RemoteSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.guard(); err != nil {
		return err
	}
	cp := r.Clone()
	if err := cp.Validate(); err != nil {
		return err
	}
	m.remote = cp
	return nil
}

// Remote 返回远端配置拷贝。
func (m *Mirror) Remote() RemoteSpec {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.remote.Clone()
}

// SetTransport 注入 Transport（测试可用 MemoryTransport）。
func (m *Mirror) SetTransport(t fetch.Transport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.guard(); err != nil {
		return err
	}
	m.transport = t
	return nil
}

// Sync 使用 Background 上下文同步。
func (m *Mirror) Sync() (*SyncReport, error) {
	return m.SyncContext(context.Background())
}

// SyncContext 拉取远端 refs/pack 并提交本地 tip。
func (m *Mirror) SyncContext(ctx context.Context) (*SyncReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrCanceled
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.guard(); err != nil {
		return nil, err
	}
	// Sync 使用远端快照，避免调用方事后改 ExtraHeaders 污染进行中的 Fetch。
	remoteSnap := m.remote.Clone()
	tr := m.transport
	if tr == nil {
		return nil, ErrNoTransport
	}
	started := time.Now().UTC()
	m.met.Inc("sync_begin")

	if err := m.lock.Wait(ctx); err != nil {
		m.met.Inc("lock_fail")
		return nil, err
	}
	m.met.Inc("lock_ok")

	rep, err := m.fetchAndApply(ctx, remoteSnap, tr)
	if err != nil {
		m.met.SetLastError(err.Error())
		_ = m.audit.Append("sync_fail", err.Error())
		return nil, err
	}
	rep.StartedAt = started
	rep.FinishedAt = time.Now().UTC()
	rep.RemoteURL = remoteSnap.URL
	m.met.Inc("sync_ok")
	m.met.Add("pack_bytes", uint64(rep.PackBytes))
	m.met.Add("ref_updates", uint64(len(rep.Updates)))
	m.met.SetLastSync(rep.FinishedAt)
	m.state.LastRemote = remoteSnap.URL
	m.state.LastSyncAt = rep.FinishedAt
	m.state.LastPackBytes = rep.PackBytes
	_ = persist.SaveSyncState(m.root, m.state)
	_ = m.audit.Append("sync_ok", fmt.Sprintf("updates=%d pack=%d", len(rep.Updates), rep.PackBytes))
	return rep, nil
}

func (m *Mirror) fetchAndApply(ctx context.Context, remote RemoteSpec, tr fetch.Transport) (*SyncReport, error) {
	have := m.refs.TipsMap()
	req := fetch.Request{
		URL:          remote.URL,
		WantPatterns: clone.Strings(remote.FetchRefs),
		Have:         have,
		Headers:      remote.HeaderMap(),
		Auth: fetch.Auth{
			Token: remote.Auth.Token,
			User:  remote.Auth.User,
			Pass:  remote.Auth.Pass,
		},
		Depth: remote.Depth,
	}
	if len(req.WantPatterns) == 0 {
		req.WantPatterns = []string{"refs/heads/*", "refs/tags/*"}
	}

	fr, err := tr.Fetch(context.Background(), req)
	if err != nil {
		if errs.IsAuth(err) {
			return nil, fmt.Errorf("%w: %v", ErrAuthFailed, err)
		}
		if ctx.Err() != nil {
			return nil, ErrCanceled
		}
		return nil, err
	}

	var packBytes int64
	var packCount int
	for i, blob := range fr.Packs {
		n, err := pack.ReceiveAndVerify(context.Background(), m.root, blob, i)
		if err != nil {
			return nil, err
		}
		packBytes += n
		packCount++
	}

	txn := m.refs.Begin()
	updates := make([]RefUpdate, 0, len(fr.Tips))
	fetched := make(map[string]ObjectID, len(fr.Tips))
	for name, oid := range fr.Tips {
		if err := validate.RefName(name); err != nil {
			txn.Abort()
			return nil, err
		}
		// Begin 已持有 store 写锁，不可再调 Resolve（会死锁）；用同步前快照。
		old := have[name]
		newOID := ObjectID(oid)
		if err := txn.Update(name, old, string(newOID), false); err != nil {
			txn.Abort()
			return nil, err
		}
		updates = append(updates, RefUpdate{
			Name:   name,
			OldOID: ObjectID(old),
			NewOID: newOID,
		})
		fetched[name] = newOID
	}

	var pruned []string
	if remote.Prune {
		pruned, err = txn.PruneAbsent(fr.Tips)
		if err != nil {
			txn.Abort()
			return nil, err
		}
	}

	// 仅在 commit 成功后内存 tip 才会切换；失败不得更新。
	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPersist, err)
	}

	return &SyncReport{
		Updates:     updates,
		PackBytes:   packBytes,
		PackCount:   packCount,
		PrunedRefs:  pruned,
		FetchedTips: fetched,
	}, nil
}

// Refs 返回本地 tip 快照（独立切片，修改不影响内部表）。
func (m *Mirror) Refs() ([]RefUpdate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.guard(); err != nil {
		return nil, err
	}
	tips := m.refs.List()
	out := make([]RefUpdate, len(tips))
	for i, t := range tips {
		out[i] = RefUpdate{
			Name:   t.Name,
			OldOID: ZeroOID,
			NewOID: ObjectID(t.OID),
		}
	}
	return out, nil
}

// Resolve 查询单 ref。
func (m *Mirror) Resolve(name string) (ObjectID, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.guard(); err != nil {
		return "", err
	}
	if err := validate.RefName(name); err != nil {
		return "", err
	}
	oid, ok := m.refs.Resolve(name)
	if !ok {
		return "", ErrNotMirrored
	}
	return ObjectID(oid), nil
}

// Prune 删除本地存在但远端 tip 集中不存在的 ref（需先 Sync 或传入 keep）。
func (m *Mirror) Prune(keep map[string]ObjectID) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.guard(); err != nil {
		return nil, err
	}
	if err := m.lock.Wait(context.Background()); err != nil {
		return nil, err
	}
	strKeep := make(map[string]string, len(keep))
	for k, v := range keep {
		strKeep[k] = string(v)
	}
	txn := m.refs.Begin()
	pruned, err := txn.PruneAbsent(strKeep)
	if err != nil {
		txn.Abort()
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return pruned, nil
}

// Stats 返回计数拷贝。
func (m *Mirror) Stats() MirrorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.met.Snapshot()
	rc := 0
	if m.refs != nil {
		rc = m.refs.Count()
	}
	return MirrorStats{
		Syncs:      s.SyncOK,
		Fetches:    s.SyncBegin,
		PackBytes:  s.PackBytes,
		RefUpdates: s.RefUpdates,
		LockWaits:  s.LockOK,
		LastSyncAt: s.LastSyncAt,
		LastError:  s.LastError,
		RefCount:   rc,
	}
}

// Root 返回镜像根目录。
func (m *Mirror) Root() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.root
}
