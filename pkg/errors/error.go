package errors

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
)

type ErrorCode int

const (
	Unknown ErrorCode = iota + 1
	NotFound
	InvalidArgument
	Internal
	Unauthenticated
	Unauthorized
	Conflict
	QuotaExceeded
)

type DetailsFunc func(st *status.Status, details ...protoiface.MessageV1) (*status.Status, error)

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

func (e *Error) GRPCCode() codes.Code {
	switch e.ErrCode {
	case NotFound:
		return codes.NotFound
	case InvalidArgument:
		return codes.InvalidArgument
	case Internal:
		return codes.Internal
	case Unauthenticated:
		return codes.Unauthenticated
	case Unauthorized:
		return codes.PermissionDenied
	case Conflict:
		return codes.AlreadyExists
	case QuotaExceeded:
		return codes.ResourceExhausted
	default:
		return codes.Unknown
	}
}

func (e *Error) GRPCStatus() *status.Status {
	st := status.New(e.GRPCCode(), e.Message)

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

func addErrorDetails(st *status.Status, details ...protoiface.MessageV1) (*status.Status, error) {
	return st.WithDetails(details...)
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
		ErrorDetailsFunc: addErrorDetails,
	}
}

func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
	return WrapError(nil, code, format, a...)
}
