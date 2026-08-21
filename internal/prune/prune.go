package prune

// Diff 计算本地有而 keep 无的 ref 名。
func Diff(local, keep map[string]string) []string {
	var out []string
	for name := range local {
		if _, ok := keep[name]; !ok {
			out = append(out, name)
		}
	}
	return out
}

// StaleTags 找出 tip 不在 keepOID 集合中的 tag ref。
func StaleTags(tips map[string]string, keepOID map[string]struct{}) []string {
	var out []string
	for name, oid := range tips {
		if !stringsHasPrefix(name, "refs/tags/") {
			continue
		}
		if _, ok := keepOID[oid]; !ok {
			out = append(out, name)
		}
	}
	return out
}

func stringsHasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}
