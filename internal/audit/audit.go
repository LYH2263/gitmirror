package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Log 镜像审计。
type Log struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

// Open 创建 audit.log。
func Open(root string) (*Log, error) {
	p := filepath.Join(root, "audit.log")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Log{path: p, f: f}, nil
}

// Append 写事件。
func (l *Log) Append(event, detail string) error {
	if l == nil || l.f == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	rec := map[string]any{
		"at":     time.Now().UTC().Format(time.RFC3339Nano),
		"event":  event,
		"detail": detail,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = l.f.Write(append(b, '\n'))
	return err
}

// Close 关闭。
func (l *Log) Close() error {
	if l == nil || l.f == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	err := l.f.Close()
	l.f = nil
	return err
}
