package pack

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/LYH2263/go-gitmirror/internal/errs"
)

const (
	magic    = "PACK"
	version  = 2
	minSize  = 12 + 20 // header + trailing sha1
)

// ReceiveAndVerify 将 pack 写入临时文件、校验后 Rename 进 objects/pack。
// 必须先关闭临时文件再 Rename（Windows 上未关闭会导致失败）。
func ReceiveAndVerify(ctx context.Context, root string, data []byte, idx int) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, errs.ErrCanceled
	}
	if err := Verify(data); err != nil {
		return 0, err
	}
	dir := filepath.Join(root, "objects", "pack")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}
	sum := sha1.Sum(data)
	name := fmt.Sprintf("pack-%s-%02d.pack", hex.EncodeToString(sum[:8]), idx)
	final := filepath.Join(dir, name)
	tmp := final + ".tmp"

	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, err
	}
	wrote, err := io.Copy(f, bytes.NewReader(data))
	if err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return 0, err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return 0, err
	}
	// 关键：先 Close 再 Rename。
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(tmp)
		return 0, errs.ErrCanceled
	}
	if err := os.Rename(tmp, final); err != nil {
		_ = os.Remove(tmp)
		return 0, err
	}
	return wrote, nil
}

// Verify 检查 PACK 头与尾部 checksum。
func Verify(data []byte) error {
	if len(data) < minSize {
		return fmt.Errorf("%w: too short", errs.ErrPackCorrupt)
	}
	if string(data[:4]) != magic {
		return fmt.Errorf("%w: bad magic", errs.ErrPackCorrupt)
	}
	ver := binary.BigEndian.Uint32(data[4:8])
	if ver != version && ver != 3 {
		return fmt.Errorf("%w: bad version %d", errs.ErrPackCorrupt, ver)
	}
	nObj := binary.BigEndian.Uint32(data[8:12])
	_ = nObj
	body := data[:len(data)-20]
	want := data[len(data)-20:]
	sum := sha1.Sum(body)
	if !bytes.Equal(sum[:], want) {
		return fmt.Errorf("%w: checksum mismatch", errs.ErrPackCorrupt)
	}
	return nil
}

// BuildMinimal 构造最小合法 pack（0 objects），供测试/MemoryTransport。
func BuildMinimal() []byte {
	hdr := make([]byte, 12)
	copy(hdr[0:4], magic)
	binary.BigEndian.PutUint32(hdr[4:8], version)
	binary.BigEndian.PutUint32(hdr[8:12], 0)
	sum := sha1.Sum(hdr)
	return append(hdr, sum[:]...)
}
