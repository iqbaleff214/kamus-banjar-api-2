package domain

import "errors"

var (
	ErrInvalidAuditAction = errors.New("invalid audit action")
	ErrCannotBanAdmin     = errors.New("cannot ban another admin")
	ErrCannotSelfDemote   = errors.New("admin cannot change their own role")
)
