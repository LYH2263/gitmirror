package gitmirror

// Close 先将内存 tip 快照落盘，再释放镜像锁与底层资源。
// 之后 Sync/Refs 返回 ErrClosed。
func (m *Mirror) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	var first error

	// Sync 仅把 tip 切到内存（Txn.Commit 不落盘），故 Close 必须在
	// 释放镜像锁之前把 tips 快照写回 tips.json 与 refs 文件，否则重开
	// 同一目录时 Store.load 读到的仍是旧快照，tip 全丢。
	if m.refs != nil {
		if err := m.refs.FlushSnapshot(); err != nil && first == nil {
			first = err
		}
	}
	if m.lock != nil {
		if err := m.lock.Unlock(); err != nil && first == nil {
			first = err
		}
	}
	if m.audit != nil {
		_ = m.audit.Close()
	}
	m.closed = true
	return first
}

func (m *Mirror) guard() error {
	if m == nil || m.closed {
		return ErrClosed
	}
	if m.lock == nil || m.refs == nil {
		return ErrClosed
	}
	return nil
}
