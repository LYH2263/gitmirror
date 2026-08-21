package remoteconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// File 简易 git-config 风格 remote 段。
type File struct {
	Name string
	URL  string
	Fetch []string
}

// Load 从 mirror 根读 remote.ini。
func Load(root string) (*File, error) {
	p := filepath.Join(root, "remote.ini")
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &File{Name: "origin"}, nil
		}
		return nil, err
	}
	defer f.Close()
	out := &File{Name: "origin"}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		switch k {
		case "name":
			out.Name = v
		case "url":
			out.URL = v
		case "fetch":
			out.Fetch = append(out.Fetch, v)
		}
	}
	return out, sc.Err()
}

// Save 写 remote.ini。
func Save(root string, cfg *File) error {
	if cfg == nil {
		return nil
	}
	var b strings.Builder
	b.WriteString("name=" + cfg.Name + "\n")
	b.WriteString("url=" + cfg.URL + "\n")
	for _, f := range cfg.Fetch {
		b.WriteString("fetch=" + f + "\n")
	}
	return os.WriteFile(filepath.Join(root, "remote.ini"), []byte(b.String()), 0o644)
}
