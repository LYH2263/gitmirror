package fetch

import (
	"bytes"
	"fmt"
	"strconv"
)

// EncodePktLine 编码一条 git pkt-line（4 hex 长度含自身）。
func EncodePktLine(payload []byte) []byte {
	n := len(payload) + 4
	if n > 65520 {
		n = 65520
		payload = payload[:65516]
	}
	head := fmt.Sprintf("%04x", n)
	out := make([]byte, 0, n)
	out = append(out, head...)
	out = append(out, payload...)
	return out
}

// EncodePktFlush 返回 flush-pkt（0000）。
func EncodePktFlush() []byte {
	return []byte("0000")
}

// DecodePktLine 解析下一条 pkt-line；flush 时 payload 为 nil 且 flush=true。
func DecodePktLine(buf []byte) (payload []byte, rest []byte, flush bool, err error) {
	if len(buf) < 4 {
		return nil, buf, false, fmt.Errorf("pkt-line: short header")
	}
	n64, err := strconv.ParseUint(string(buf[:4]), 16, 16)
	if err != nil {
		return nil, buf, false, fmt.Errorf("pkt-line: bad len: %w", err)
	}
	if n64 == 0 {
		return nil, buf[4:], true, nil
	}
	n := int(n64)
	if n < 4 || len(buf) < n {
		return nil, buf, false, fmt.Errorf("pkt-line: truncated want %d have %d", n, len(buf))
	}
	return buf[4:n], buf[n:], false, nil
}

// SplitPktLines 拆分完整流为 payload 列表（不含 flush）。
func SplitPktLines(stream []byte) ([][]byte, error) {
	var out [][]byte
	rest := stream
	for len(rest) > 0 {
		p, next, flush, err := DecodePktLine(rest)
		if err != nil {
			return nil, err
		}
		rest = next
		if flush {
			continue
		}
		// 去掉行尾 LF（若有）
		p = bytes.TrimSuffix(p, []byte{'\n'})
		out = append(out, p)
	}
	return out, nil
}

// BuildWantHave 构造简化的 want/have 协商正文（测试/内存 Transport 可用）。
func BuildWantHave(wants, haves []string) []byte {
	var b bytes.Buffer
	for i, w := range wants {
		line := "want " + w
		if i == 0 {
			line += " multi_ack_detailed no-done agent=go-gitmirror"
		}
		b.Write(EncodePktLine([]byte(line + "\n")))
	}
	b.Write(EncodePktFlush())
	for _, h := range haves {
		b.Write(EncodePktLine([]byte("have " + h + "\n")))
	}
	b.Write(EncodePktLine([]byte("done\n")))
	return b.Bytes()
}
