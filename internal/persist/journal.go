package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Journal 追加式操作日志。
type Journal struct {
	mu   sync.Mutex
	path string
}

// OpenJournal 打开。
func OpenJournal(root string) (*Journal, error) {
	p := filepath.Join(root, "journal.jsonl")
	f, err := os.OpenFile(p, os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	_ = f.Close()
	return &Journal{path: p}, nil
}

// Entry 日志行。
type Entry struct {
	At   time.Time `json:"at"`
	Op   string    `json:"op"`
	Detail string  `json:"detail"`
}

// Append 写一行。
func (j *Journal) Append(op, detail string) error {
	if j == nil {
		return nil
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	e := Entry{At: time.Now().UTC(), Op: op, Detail: detail}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(j.path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}
