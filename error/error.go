package error

type InternalErrorCode int

const ErrorCodeInternalErrorCodeSomething InternalErrorCode = 0

type Error interface {
	InternalErrorCode() InternalErrorCode
	Error() string
}
