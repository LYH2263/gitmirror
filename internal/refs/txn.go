package refs

import (
	"fmt"

	"github.com/LYH2263/go-gitmirror/internal/errs"
)

// Txn ref 更新事务：Commit 成功才写回 store.tips。
type Txn struct {
	store   *Store
	pending map[string]string
	deleted []string
	active  bool
}

// Update 暂存 tip 变更；conflict 时返回 ErrRefConflict。
func (t *Txn) Update(name, oldOID, newOID string, force bool) error {
	if t == nil || !t.active {
		return errs.ErrClosed
	}
	cur, ok := t.pending[name]
	if !ok {
		cur = ""
	}
	if !force && oldOID != "" && cur != "" && cur != oldOID {
		return fmt.Errorf("%w: %s have %s want old %s", errs.ErrRefConflict, name, cur, oldOID)
	}
	t.pending[name] = newOID
	return nil
}

// Delete 标记删除。
func (t *Txn) Delete(name string) error {
	if t == nil || !t.active {
		return errs.ErrClosed
	}
	delete(t.pending, name)
	t.deleted = append(t.deleted, name)
	return nil
}

// PruneAbsent 删除不在 keep 中的本地 tip。
func (t *Txn) PruneAbsent(keep map[string]string) ([]string, error) {
	if t == nil || !t.active {
		return nil, errs.ErrClosed
	}
	var pruned []string
	for name := range t.pending {
		if _, ok := keep[name]; !ok {
			pruned = append(pruned, name)
		}
	}
	for _, name := range pruned {
		delete(t.pending, name)
		t.deleted = append(t.deleted, name)
	}
	return pruned, nil
}

// Commit 落盘并切换内存 tip。失败时不得更新 store.tips。
func (t *Txn) Commit() error {
	if t == nil || !t.active {
		return errs.ErrClosed
	}

	t.store.tips = t.pending
	t.active = false
	t.store.mu.Unlock()
	return nil
}

// Abort 丢弃。
func (t *Txn) Abort() {
	if t == nil || !t.active {
		return
	}
	t.active = false
	t.store.mu.Unlock()
}
