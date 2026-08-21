package object

import (
	"crypto/sha1"
	"fmt"
)

// Type Git 对象类型。
type Type byte

const (
	TypeBad    Type = 0
	TypeCommit Type = 1
	TypeTree   Type = 2
	TypeBlob   Type = 3
	TypeTag    Type = 4
)

func (t Type) String() string {
	switch t {
	case TypeCommit:
		return "commit"
	case TypeTree:
		return "tree"
	case TypeBlob:
		return "blob"
	case TypeTag:
		return "tag"
	default:
		return "bad"
	}
}

// Header 生成松散对象存储头 "type size\0"。
func Header(t Type, size int) []byte {
	return []byte(fmt.Sprintf("%s %d\x00", t.String(), size))
}

// TreeEntry 树对象条目。
type TreeEntry struct {
	Mode string
	Name string
	OID  ID
}

// EncodeTree 编码 tree 对象内容（不含 header）。
func EncodeTree(ents []TreeEntry) []byte {
	var out []byte
	for _, e := range ents {
		out = append(out, []byte(e.Mode)...)
		out = append(out, ' ')
		out = append(out, []byte(e.Name)...)
		out = append(out, 0)
		out = append(out, e.OID[:]...)
	}
	return out
}

// HashTree 计算 tree OID。
func HashTree(ents []TreeEntry) ID {
	body := EncodeTree(ents)
	h := sha1.New()
	_, _ = fmt.Fprintf(h, "tree %d\x00", len(body))
	_, _ = h.Write(body)
	var id ID
	copy(id[:], h.Sum(nil))
	return id
}

// HashCommit 计算 commit OID（raw 正文，不含 header）。
func HashCommit(body []byte) ID {
	h := sha1.New()
	_, _ = fmt.Fprintf(h, "commit %d\x00", len(body))
	_, _ = h.Write(body)
	var id ID
	copy(id[:], h.Sum(nil))
	return id
}

// FormatCommit 构造极简 commit 正文。
func FormatCommit(tree ID, parents []ID, author, message string) []byte {
	var b []byte
	b = append(b, []byte("tree "+tree.String()+"\n")...)
	for _, p := range parents {
		b = append(b, []byte("parent "+p.String()+"\n")...)
	}
	b = append(b, []byte("author "+author+"\n")...)
	b = append(b, []byte("committer "+author+"\n")...)
	b = append(b, '\n')
	b = append(b, []byte(message)...)
	if len(message) == 0 || message[len(message)-1] != '\n' {
		b = append(b, '\n')
	}
	return b
}
