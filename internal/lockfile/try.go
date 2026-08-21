package lockfile

import (
	"os"

	"github.com/LYH2263/go-gitmirror/internal/errs"
)

// TryLock 非阻塞尝试；失败返回 ErrLockHeld。
func TryLock(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, errs.ErrLockHeld
		}
		return nil, err
	}
	return f, nil
}
