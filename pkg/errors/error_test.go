package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
)

func TestError(t *testing.T) {

	t.Run("Error", func(t *testing.T) {
		err := WrapError(nil, Internal, "error ocurred")
		assert.Equal(t, err.Error(), "code: 4, message: error ocurred")

		orig := errors.New("original error")
		err = WrapError(orig, Internal, "wrapped error")

		assert.NotNil(t, err)

		customErr, ok := err.(*Error)
		assert.True(t, ok)
		assert.Equal(t, customErr.Error(), "code: 4, message: wrapped error, original error: original error")
	})

	t.Run("Code", func(t *testing.T) {
		orig := errors.New("original error")
		err := WrapError(orig, Internal, "wrapped error")

		assert.NotNil(t, err)

		customErr, _ := err.(*Error)
		assert.Equal(t, customErr.Code(), Internal)
	})

	t.Run("Unwrap", func(t *testing.T) {
		orig := errors.New("original error")
		err := WrapError(orig, Internal, "wrapped error")

		assert.True(t, errors.Is(err, orig))
	})

	t.Run("With Violations", func(t *testing.T) {
		v := []*FieldViolation{
			{"name", "required"},
		}
		err := NewValidationError(v, "bad request")
		require.Error(t, err)
		assert.Equal(t, 1, len(err.FieldViolations))
		assert.Equal(t, err.Error(), "validation error(s): [name: required]")
	})
}

func TestHTTPCodeMappings(t *testing.T) {
	cases := map[ErrorCode]int{
		NotFound:             http.StatusNotFound,
		InvalidArgument:      http.StatusBadRequest,
		BadRequest:           http.StatusBadRequest,
		Internal:             http.StatusInternalServerError,
		FailedDependency:     http.StatusInternalServerError,
		Unauthenticated:      http.StatusUnauthorized,
		Unauthorized:         http.StatusForbidden,
		Conflict:             http.StatusConflict,
		PreconditionFailed:   http.StatusConflict,
		QuotaExceeded:        http.StatusTooManyRequests,
		NotImplemented:       http.StatusNotImplemented,
		ServiceUnavailable:   http.StatusServiceUnavailable,
		DeadlineExceeded:     http.StatusGatewayTimeout,
		TooEarly:             http.StatusTooEarly,
		Gone:                 http.StatusGone,
		UnprocessableEntity:  http.StatusUnprocessableEntity,
		PayloadTooLarge:      http.StatusRequestEntityTooLarge,
		UnsupportedMediaType: http.StatusUnsupportedMediaType,
		Unknown:              http.StatusInternalServerError,
	}

	for code, expected := range cases {
		e := &Error{ErrCode: code}
		assert.Equal(t, expected, e.HTTPCode(), "HTTPCode(%v) mismatch", code)
	}
}

func TestGRPCCodeMappings(t *testing.T) {
	cases := map[ErrorCode]codes.Code{
		NotFound:             codes.NotFound,
		InvalidArgument:      codes.InvalidArgument,
		BadRequest:           codes.InvalidArgument,
		UnprocessableEntity:  codes.InvalidArgument,
		Internal:             codes.Internal,
		FailedDependency:     codes.Internal,
		Unauthenticated:      codes.Unauthenticated,
		Unauthorized:         codes.PermissionDenied,
		Conflict:             codes.Aborted,
		PreconditionFailed:   codes.Aborted,
		QuotaExceeded:        codes.ResourceExhausted,
		NotImplemented:       codes.Unimplemented,
		UnsupportedMediaType: codes.Unimplemented,
		ServiceUnavailable:   codes.Unavailable,
		DeadlineExceeded:     codes.DeadlineExceeded,
		TooEarly:             codes.DeadlineExceeded,
		Gone:                 codes.NotFound,
		PayloadTooLarge:      codes.OutOfRange,
		Unknown:              codes.Unknown,
	}

	for code, expected := range cases {
		e := &Error{ErrCode: code}
		assert.Equal(t, expected, e.GRPCCode(), "GRPCCode(%v) mismatch", code)
	}
}

func TestHTTPStatus(t *testing.T) {

	t.Run("Include Details", func(t *testing.T) {
		e := &Error{
			Message: "something bad",
			ErrCode: InvalidArgument,
			FieldViolations: []*FieldViolation{
				{"name", "required"},
			},
			Original: errors.New("invalid input"),
		}

		code, body := e.HTTPStatus()
		assert.Equal(t, http.StatusBadRequest, code)
		assert.Equal(t, e.Message, body["message"])
		assert.Equal(t, e.Original.Error(), body["reason"])

		violations, ok := body["violations"].([]map[string]string)
		require.True(t, ok, "violations key should exist")
		require.Len(t, violations, 1)
		assert.Equal(t, "name", violations[0]["field"])
		assert.Equal(t, "required", violations[0]["description"])
	})
}

func TestGRPCStatus(t *testing.T) {

	t.Run("Wrap Existing Error", func(t *testing.T) {
		err := &Error{
			Original: errors.New("panic"),
			Message:  "test",
			ErrCode:  Internal,
			ErrorDetailsFunc: func(st *status.Status, _ ...protoiface.MessageV1) (*status.Status, error) {
				return nil, errors.New("fail")
			},
		}
		st := err.GRPCStatus()
		assert.NotNil(t, st, "expected fallback status even if detail func fails")
	})

	t.Run("Details Added", func(t *testing.T) {
		err := &Error{
			Original: errors.New("panic"),
			Message:  "test",
			ErrCode:  Internal,
			ErrorDetailsFunc: func(st *status.Status, _ ...protoiface.MessageV1) (*status.Status, error) {
				return status.New(codes.Internal, "test"), nil
			},
		}
		st := err.GRPCStatus()
		assert.NotNil(t, st, "expected fallback status even if detail func fails")
	})

	t.Run("New Error", func(t *testing.T) {
		e := &Error{
			Message:          "internal",
			ErrCode:          Internal,
			ErrorDetailsFunc: addErrorDetails,
		}
		st := e.GRPCStatus()
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("With Violations", func(t *testing.T) {
		err := &Error{
			Message:          "bad input",
			ErrCode:          InvalidArgument,
			ErrorDetailsFunc: addErrorDetails,
			FieldViolations: []*FieldViolation{
				{"age", "must be > 0"},
			},
		}
		st := err.GRPCStatus()
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("With Violations Details Failure", func(t *testing.T) {
		err := &Error{
			Message: "bad input",
			ErrCode: InvalidArgument,
			FieldViolations: []*FieldViolation{
				{"age", "must be > 0"},
			},
			ErrorDetailsFunc: func(st *status.Status, _ ...protoiface.MessageV1) (*status.Status, error) {
				return nil, errors.New("fail")
			},
		}
		st := err.GRPCStatus()
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestNewErrorHelpers(t *testing.T) {
	t.Run("Wrap Existing Error", func(t *testing.T) {
		err := WrapError(errors.New("db"), Internal, "failed: %s", "query")
		_, ok := err.(*Error)
		require.True(t, ok)
	})

	t.Run("New Error", func(t *testing.T) {
		err := NewErrorf(NotFound, "missing %s", "record")
		cerr := err.(*Error)
		assert.Equal(t, "missing record", cerr.Message)
	})
}

func TestBuildViolations(t *testing.T) {
	validate := validator.New()
	type Input struct {
		Name string `validate:"required"`
	}
	input := Input{}
	err := validate.Struct(input)
	require.Error(t, err)

	verrs := err.(validator.ValidationErrors)
	v := BuildViolations(verrs)
	require.NotEmpty(t, v)
	assert.Equal(t, "Name", v[0].Field)
}

func TestNewValidationError(t *testing.T) {
	validate := validator.New()
	type Input struct {
		Name string `validate:"required"`
	}
	input := Input{}
	err := validate.Struct(input)
	require.Error(t, err)

	verrs := err.(validator.ValidationErrors)
	e := NewValidationError(BuildViolations(verrs), "bad request")
	assert.Equal(t, InvalidArgument, e.ErrCode)
	assert.Equal(t, "bad request", e.Message)
	assert.Len(t, e.FieldViolations, 1)
}
