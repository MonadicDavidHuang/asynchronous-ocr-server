package repository

import (
	ettot "asynchronous-ocr-server/error"
	"fmt"
)

// [1, 9999] are reserved for internal error code
const (
	internalErrorCodeBase               ettot.InternalErrorCode = 0
	InternalErrorCodeSystemRelatedError ettot.InternalErrorCode = iota + 1 + internalErrorCodeBase
	InternalErrorCodeNoRecordFoundError
)

var errorMap = map[ettot.InternalErrorCode]string{
	InternalErrorCodeSystemRelatedError: "db query failed due to system related error, child-error(%w)",
	InternalErrorCodeNoRecordFoundError: "no record found, child-error(%w)",
}

// RepositoryError ...,
type RepositoryError struct {
	Err               error
	internalErrorCode ettot.InternalErrorCode
}

func (e *RepositoryError) InternalErrorCode() ettot.InternalErrorCode {
	return e.internalErrorCode
}

func (e *RepositoryError) Error() string {
	msg := errorMap[e.internalErrorCode]
	return fmt.Errorf(msg, e.Err).Error()
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

func NewNRepositoryError(
	err error,
	internalErrorCode ettot.InternalErrorCode,
) ettot.Error {
	return &RepositoryError{
		err,
		internalErrorCode,
	}
}
