package service

import (
	ettot "asynchronous-ocr-server/error"
	"fmt"
)

// [10001, 19999] are reserved for internal error code
const (
	internalErrorCodeBase               ettot.InternalErrorCode = 10000
	InternalErrorCodeSystemRelatedError ettot.InternalErrorCode = iota + 1 + internalErrorCodeBase
	InternalErrorCodeNoTaskFoundError
	InternalErrorCodeTaskIsPendingError
	InternalErrorCodeTaskIsDeletedError
	InternalErrorCodeNoImageFileFoundError
	InternalErrorCodeFailedToStoreImageFileError
	InternalErrorCodeFailedToCreateTaskError
	InternalErrorCodeFailedToGetTaskError
	InternalErrorCodeFailedToDeleteTaskError
	InternalErrorCodeFailedToApplyOCRError
)

var errorMap = map[ettot.InternalErrorCode]string{
	InternalErrorCodeSystemRelatedError:          "service failed to retrieve information due to system related error(%w)",
	InternalErrorCodeNoTaskFoundError:            "no task found, child-error(%w)",
	InternalErrorCodeNoImageFileFoundError:       "no image file found, child-error(%w)",
	InternalErrorCodeFailedToStoreImageFileError: "failed to store image file, child-error(%w)",
	InternalErrorCodeFailedToCreateTaskError:     "failed to create task, child-error(%w)",
	InternalErrorCodeFailedToGetTaskError:        "failed to get task, child-error(%w)",
	InternalErrorCodeFailedToDeleteTaskError:     "failed to delete task, child-error(%w)",
	InternalErrorCodeFailedToApplyOCRError:       "failed to apply OCR, child-error(%w)",
	InternalErrorCodeTaskIsPendingError:          "queried task is still pending, child-error(%w)",
	InternalErrorCodeTaskIsDeletedError:          "queried task is alreayd deleted, , child-error(%w)",
}

// ServiceError ..., TODO: using better naming
type ServiceError struct {
	Err               error
	internalErrorCode ettot.InternalErrorCode
}

func (e *ServiceError) InternalErrorCode() ettot.InternalErrorCode {
	return e.internalErrorCode
}

func (e *ServiceError) Error() string {
	msg := errorMap[e.internalErrorCode]
	return fmt.Errorf(msg, e.Err).Error()
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func NewServiceError(
	err error,
	internalErrorCode ettot.InternalErrorCode,
) ettot.Error {
	return &ServiceError{
		err,
		internalErrorCode,
	}
}
