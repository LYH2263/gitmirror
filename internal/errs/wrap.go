package errs

import "errors"

// IsAuth 判断是否认证类错误（含包装）。
func IsAuth(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrAuthFailed) {
		return true
	}
	var ae *AuthError
	return errors.As(err, &ae)
}

// AuthError Transport 层认证失败。
type AuthError struct {
	Msg string
}

func (e *AuthError) Error() string {
	if e.Msg == "" {
		return ErrAuthFailed.Error()
	}
	return e.Msg
}

func (e *AuthError) Unwrap() error { return ErrAuthFailed }

// AsAuth 构造可 %w 的认证错误。
func AsAuth(msg string) error {
	return &AuthError{Msg: msg}
}
