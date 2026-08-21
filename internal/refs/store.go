package refs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/LYH2263/go-gitmirror/internal/clone"
)

// Tip 本地 ref tip。
type Tip struct {
	Name string `json:"name"`
	OID  string `json:"oid"`
}

// Store 文件系统模拟的 refs 表（refs/heads/... 文件 + tips.json 快照）。
type Store struct {
	mu   sync.RWMutex
	root string
	tips map[string]string
}

// Open 加载或初始化。
func Open(root string) (*Store, error) {
	refRoot := filepath.Join(root, "refs")
	if err := os.MkdirAll(refRoot, 0o755); err != nil {
		return nil, err
	}
	_ = os.MkdirAll(filepath.Join(root, "objects", "pack"), 0o755)
	s := &Store{
		root: root,
		tips: make(map[string]string),
	}
	if err := LoadPackedRefs(root, s.tips); err != nil {
		return nil, err
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) snapshotPath() string {
	return filepath.Join(s.root, "tips.json")
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.snapshotPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var tips []Tip
	if err := json.Unmarshal(b, &tips); err != nil {
		return err
	}
	for _, t := range tips {
		s.tips[t.Name] = t.OID
	}
	return s.walkRefFiles(filepath.Join(s.root, "refs"), "refs")
}

func (s *Store) walkRefFiles(dir, prefix string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		name := prefix + "/" + e.Name()
		if e.IsDir() {
			if err := s.walkRefFiles(p, name); err != nil {
				return err
			}
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		oid := string(bytesTrim(b))
		if oid != "" {
			s.tips[name] = oid
		}
	}
	return nil
}

func bytesTrim(b []byte) []byte {
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n' || b[i] == '\r') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\t' || b[j-1] == '\n' || b[j-1] == '\r') {
		j--
	}
	return b[i:j]
}

// FlushSnapshot 将内存 tip 写入 tips.json 与 refs 文件。
func (s *Store) FlushSnapshot() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushLocked()
}

func (s *Store) flushLocked() error {
	tips := make([]Tip, 0, len(s.tips))
	for n, o := range s.tips {
		tips = append(tips, Tip{Name: n, OID: o})
	}
	sort.Slice(tips, func(i, j int) bool { return tips[i].Name < tips[j].Name })
	b, err := json.MarshalIndent(tips, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.snapshotPath() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.snapshotPath()); err != nil {
		return err
	}
	for _, t := range tips {
		if err := writeRefFile(s.root, t.Name, t.OID); err != nil {
			return err
		}
	}
	return nil
}

func writeRefFile(root, name, oid string) error {
	p := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(oid+"\n"), 0o644)
}

// Resolve 查询。
func (s *Store) Resolve(name string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	oid, ok := s.tips[name]
	return oid, ok
}

// List 返回 tip 拷贝切片（每次独立分配，修改不影响内部表）。
func (s *Store) List() []Tip {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Tip, 0, len(s.tips))
	for n, o := range s.tips {
		out = append(out, Tip{Name: n, OID: o})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// TipsMap 返回 OID map 拷贝。
func (s *Store) TipsMap() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone.StringMap(s.tips)
}

// Count tip 数量。
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tips)
}

// Close 刷盘。
func (s *Store) Close() error {
	return s.FlushSnapshot()
}

// Begin 开启事务。
func (s *Store) Begin() *Txn {
	s.mu.Lock()
	pending := clone.StringMap(s.tips)
	if pending == nil {
		pending = make(map[string]string)
	}
	return &Txn{store: s, pending: pending, active: true}
}
