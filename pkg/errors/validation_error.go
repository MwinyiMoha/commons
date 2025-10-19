package errors

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FieldViolation struct {
	Field       string
	Description string
}

type ValidationError struct {
	FieldViolations  []*FieldViolation
	ErrorDetailsFunc DetailsFunc
}

func (e *ValidationError) Error() string {
	var violationMessages []string
	for _, v := range e.FieldViolations {
		violationMessages = append(violationMessages, fmt.Sprintf("%s: %s", v.Field, v.Description))
	}

	return fmt.Sprintf("validation error(s): %v", violationMessages)
}

func (e *ValidationError) GRPCStatus() *status.Status {
	st := status.New(codes.InvalidArgument, "bad request")

	violations := []*errdetails.BadRequest_FieldViolation{}
	for _, v := range e.FieldViolations {
		violations = append(
			violations,
			&errdetails.BadRequest_FieldViolation{
				Field:       v.Field,
				Description: v.Description,
			},
		)
	}

	detailed, err := e.ErrorDetailsFunc(st, &errdetails.BadRequest{FieldViolations: violations})
	if err != nil {
		return st
	}

	return detailed
}

func BuildViolations(verrs validator.ValidationErrors) []*FieldViolation {
	var violations []*FieldViolation

	for _, err := range verrs {
		violations = append(
			violations,
			&FieldViolation{
				Field:       err.StructField(),
				Description: err.Error(),
			},
		)
	}

	return violations
}

func NewValidationError(violations []*FieldViolation) *ValidationError {
	return &ValidationError{
		FieldViolations:  violations,
		ErrorDetailsFunc: injectDetails,
	}
}
