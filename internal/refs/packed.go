package refs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadPackedRefs 读取 packed-refs（若存在），合并进 dst（不覆盖已有键）。
func LoadPackedRefs(root string, dst map[string]string) error {
	p := filepath.Join(root, "packed-refs")
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "^") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		oid, name := parts[0], parts[1]
		if _, ok := dst[name]; !ok {
			dst[name] = oid
		}
	}
	return sc.Err()
}

// WritePackedRefs 将 tips 写成 packed-refs（peeled 行省略）。
func WritePackedRefs(root string, tips map[string]string) error {
	var b strings.Builder
	b.WriteString("# pack-refs with: peeled fully-peeled sorted\n")
	names := make([]string, 0, len(tips))
	for n := range tips {
		names = append(names, n)
	}
	// 简单插入排序，避免额外 import
	for i := 1; i < len(names); i++ {
		j := i
		for j > 0 && names[j] < names[j-1] {
			names[j], names[j-1] = names[j-1], names[j]
			j--
		}
	}
	for _, n := range names {
		b.WriteString(tips[n])
		b.WriteByte(' ')
		b.WriteString(n)
		b.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(root, "packed-refs"), []byte(b.String()), 0o644)
}
