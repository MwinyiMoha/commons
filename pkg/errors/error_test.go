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

func TestGRPCCode(t *testing.T) {

	t.Run("Internal", func(t *testing.T) {
		err := NewErrorf(Internal, "")
		assert.NotNil(t, err)
		assert.Equal(t, codes.Internal, err.(*Error).GRPCCode())
	})
	t.Run("Not found", func(t *testing.T) {
		err := NewErrorf(NotFound, "")
		assert.NotNil(t, err)
		assert.Equal(t, codes.NotFound, err.(*Error).GRPCCode())
	})

	t.Run("Bad request", func(t *testing.T) {
		err := NewErrorf(InvalidArgument, "")
		assert.NotNil(t, err)
		assert.Equal(t, codes.InvalidArgument, err.(*Error).GRPCCode())
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		err := NewErrorf(Unauthenticated, "")
		assert.NotNil(t, err)
		assert.Equal(t, codes.Unauthenticated, err.(*Error).GRPCCode())
	})

	t.Run("Unauthorized", func(t *testing.T) {
		err := NewErrorf(Unauthorized, "")
		assert.NotNil(t, err)
		assert.Equal(t, codes.PermissionDenied, err.(*Error).GRPCCode())
	})

	t.Run("Conflict", func(t *testing.T) {
		err := NewErrorf(Conflict, "")
		assert.NotNil(t, err)
		assert.Equal(t, codes.AlreadyExists, err.(*Error).GRPCCode())
	})

	t.Run("Resource exhausted", func(t *testing.T) {
		err := NewErrorf(QuotaExceeded, "")
		assert.NotNil(t, err)
		assert.Equal(t, codes.ResourceExhausted, err.(*Error).GRPCCode())
	})

	t.Run("Unknown", func(t *testing.T) {
		err := NewErrorf(Unknown, "")
		assert.NotNil(t, err)
		assert.Equal(t, codes.Unknown, err.(*Error).GRPCCode())
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

	t.Run("Error details addition", func(t *testing.T) {
		st := status.New(codes.InvalidArgument, "invalid argument")
		st, err := addErrorDetails(st, &errdetails.ErrorInfo{})

		assert.Nil(t, err)
		assert.Equal(t, st.Code(), codes.InvalidArgument)
	})
}
