package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
)

type DetailsFunc func(st *status.Status, details ...protoiface.MessageV1) (*status.Status, error)

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

func injectDetails(st *status.Status, details ...protoiface.MessageV1) (*status.Status, error) {
	return st.WithDetails(details...)
}

func toGRPCCode(code ErrorCode) codes.Code {
	switch code {
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
