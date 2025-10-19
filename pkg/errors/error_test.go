package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
)

func TestError(t *testing.T) {

	t.Run("Error", func(t *testing.T) {
		orig := fmt.Errorf("original error")
		err := WrapError(orig, Internal, "wrapped error")

		assert.NotNil(t, err)

		customErr, ok := err.(*Error)
		assert.True(t, ok)
		assert.Equal(t, customErr.Error(), "code: 4, message: wrapped error, original error: original error")
	})

	t.Run("Code", func(t *testing.T) {
		orig := fmt.Errorf("original error")
		err := WrapError(orig, Internal, "wrapped error")

		assert.NotNil(t, err)

		customErr, _ := err.(*Error)
		assert.Equal(t, customErr.Code(), Internal)
	})

	t.Run("Unwrap", func(t *testing.T) {
		orig := fmt.Errorf("original error")
		err := WrapError(orig, Internal, "wrapped error")

		assert.True(t, errors.Is(err, orig))
	})
}

func TestNewErrorf(t *testing.T) {
	t.Run("Formatted message", func(t *testing.T) {
		err := NewErrorf(NotFound, "Item %s not found", "123")
		assert.NotNil(t, err)

		customErr, ok := err.(*Error)
		assert.True(t, ok, "Expected error type *errors.Error")
		assert.Equal(t, "Item 123 not found", customErr.Message)
		assert.Equal(t, "code: 2, message: Item 123 not found", customErr.Error())
	})

	t.Run("Static message", func(t *testing.T) {
		err := NewErrorf(NotFound, "Item 123 not found")
		assert.NotNil(t, err)

		_, ok := err.(*Error)
		assert.True(t, ok, "Expected error type *errors.Error")
	})
}

func TestGRPCStatus(t *testing.T) {

	t.Run("With original error", func(t *testing.T) {
		original := errors.New("database connection failed")
		err := WrapError(original, Internal, "Failed to retrieve data")

		grpcStatus := err.(*Error).GRPCStatus()
		assert.Equal(t, codes.Internal, grpcStatus.Code())
		assert.Equal(t, "Failed to retrieve data", grpcStatus.Message())
	})

	t.Run("No original error", func(t *testing.T) {
		err := NewErrorf(Internal, "database connection failed")
		grpcStatus := err.(*Error).GRPCStatus()

		assert.Equal(t, codes.Internal, grpcStatus.Code())
		assert.Equal(t, "database connection failed", grpcStatus.Message())
	})

	t.Run("Failing details addition", func(t *testing.T) {
		mockInjectDetails := func(st *status.Status, details ...protoiface.MessageV1) (*status.Status, error) {
			return nil, fmt.Errorf("mocked inject details error")
		}

		original := fmt.Errorf("database connection failed")
		err := WrapError(original, Internal, "internal error")
		customErr, _ := err.(*Error)
		customErr.ErrorDetailsFunc = mockInjectDetails

		grpcStatus := customErr.GRPCStatus()
		assert.Equal(t, grpcStatus.Code(), codes.Internal)
		assert.Equal(t, "internal error", grpcStatus.Message())
	})
}

func TestHelpers(t *testing.T) {
	t.Run("Status code mapping", func(t *testing.T) {
		assert.Equal(t, codes.Internal, toGRPCCode(Internal))
		assert.Equal(t, codes.InvalidArgument, toGRPCCode(InvalidArgument))
		assert.Equal(t, codes.ResourceExhausted, toGRPCCode(QuotaExceeded))
		assert.Equal(t, codes.NotFound, toGRPCCode(NotFound))
		assert.Equal(t, codes.Unauthenticated, toGRPCCode(Unauthenticated))
		assert.Equal(t, codes.PermissionDenied, toGRPCCode(Unauthorized))
		assert.Equal(t, codes.AlreadyExists, toGRPCCode(Conflict))
		assert.Equal(t, codes.Unknown, toGRPCCode(Unknown))
	})

	t.Run("Error details addition", func(t *testing.T) {
		st := status.New(codes.InvalidArgument, "invalid argument")
		st, err := injectDetails(st, &errdetails.ErrorInfo{})

		assert.Nil(t, err)
		assert.Equal(t, st.Code(), codes.InvalidArgument)
	})
}
