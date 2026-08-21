package pack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// IndexEntry 已接收 pack 登记。
type IndexEntry struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	SHA  string `json:"sha"`
}

// ListPacks 列举 objects/pack 下 .pack 文件。
func ListPacks(root string) ([]IndexEntry, error) {
	dir := filepath.Join(root, "objects", "pack")
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []IndexEntry
	for _, e := range ents {
		if e.IsDir() || filepath.Ext(e.Name()) != ".pack" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, IndexEntry{Name: e.Name(), Size: info.Size()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// WriteIndex 写 pack-index.json。
func WriteIndex(root string, entries []IndexEntry) error {
	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(root, "objects", "pack", "pack-index.json")
	return os.WriteFile(p, b, 0o644)
}
