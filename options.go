package gitmirror

import (
	"time"

	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

// Options 打开 Mirror。
type Options struct {
	Root      string
	Remote    RemoteSpec
	Transport fetch.Transport
	LockWait  time.Duration // 等镜像锁上限；0 用默认 30s
	MirrorAll bool         // true 时本地命名空间与远端一致（默认）
}
