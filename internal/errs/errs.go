package errs

import "errors"

var (
	ErrClosed        = errors.New("gitmirror: closed")
	ErrNoTransport   = errors.New("gitmirror: no transport")
	ErrAuthFailed    = errors.New("gitmirror: auth failed")
	ErrInvalidRemote = errors.New("gitmirror: invalid remote")
	ErrInvalidRef    = errors.New("gitmirror: invalid ref")
	ErrLockTimeout   = errors.New("gitmirror: lock timeout")
	ErrPackCorrupt   = errors.New("gitmirror: pack corrupt")
	ErrRefConflict   = errors.New("gitmirror: ref conflict")
	ErrCanceled      = errors.New("gitmirror: canceled")
	ErrPersist       = errors.New("gitmirror: persist")
	ErrNotMirrored   = errors.New("gitmirror: ref not mirrored")
	ErrLockHeld      = errors.New("gitmirror: lock held")
	ErrBadOID        = errors.New("gitmirror: bad object id")
)
