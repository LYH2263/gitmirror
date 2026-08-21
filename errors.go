package gitmirror

import "github.com/LYH2263/go-gitmirror/internal/errs"

var (
	ErrClosed       = errs.ErrClosed
	ErrNoTransport  = errs.ErrNoTransport
	ErrAuthFailed   = errs.ErrAuthFailed
	ErrInvalidRemote = errs.ErrInvalidRemote
	ErrInvalidRef   = errs.ErrInvalidRef
	ErrLockTimeout  = errs.ErrLockTimeout
	ErrPackCorrupt  = errs.ErrPackCorrupt
	ErrRefConflict  = errs.ErrRefConflict
	ErrCanceled     = errs.ErrCanceled
	ErrPersist      = errs.ErrPersist
	ErrNotMirrored  = errs.ErrNotMirrored
)
