package fetch

import "strings"

// NegotiateWants 根据 have 与远端 advertised 计算 still-want。
func NegotiateWants(advertised, have map[string]string) map[string]string {
	want := make(map[string]string)
	for name, oid := range advertised {
		if cur, ok := have[name]; ok && cur == oid {
			continue
		}
		want[name] = oid
	}
	return want
}

// FilterNamespace 仅保留某前缀下的 tip。
func FilterNamespace(tips map[string]string, ns string) map[string]string {
	if !strings.HasSuffix(ns, "/") {
		ns += "/"
	}
	out := make(map[string]string)
	for k, v := range tips {
		if strings.HasPrefix(k, ns) {
			out[k] = v
		}
	}
	return out
}
