package errors

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type Error struct {
	Original         error
	Message          string
	ErrCode          ErrorCode
	ErrorDetailsFunc DetailsFunc
}

func (e *Error) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("code: %d, message: %s, original error: %s", e.ErrCode, e.Message, e.Original)
	}

	return fmt.Sprintf("code: %d, message: %s", e.ErrCode, e.Message)
}

func (e *Error) Code() ErrorCode {
	return e.ErrCode
}

func (e *Error) Unwrap() error {
	return e.Original
}

func (e *Error) GRPCStatus() *status.Status {
	statusCode := toGRPCCode(e.ErrCode)
	st := status.New(statusCode, e.Message)

	if e.Original != nil {
		details := &errdetails.ErrorInfo{
			Reason:   e.Original.Error(),
			Metadata: map[string]string{"custom_code": fmt.Sprintf("%d", e.ErrCode)},
		}

		detailed, err := e.ErrorDetailsFunc(st, details)
		if err != nil {
			return st
		}

		return detailed
	}

	return st
}

func WrapError(original error, code ErrorCode, format string, a ...interface{}) error {
	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}

	return &Error{
		ErrCode:          code,
		Original:         original,
		Message:          message,
		ErrorDetailsFunc: injectDetails,
	}
}

func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
	return WrapError(nil, code, format, a...)
}
