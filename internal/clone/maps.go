package clone

// StringStringMap 别名，便于调用方语义。
func StringStringMap(src map[string]string) map[string]string {
	return StringMap(src)
}

// MergeStringMap 将 b 合并进 a 的拷贝。
func MergeStringMap(a, b map[string]string) map[string]string {
	out := StringMap(a)
	if out == nil {
		out = make(map[string]string, len(b))
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
