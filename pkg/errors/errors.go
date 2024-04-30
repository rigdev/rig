package errors

import (
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	e connect.Error
}

func (e Error) Unwrap() error {
	return &e.e
}

func (e Error) Error() string {
	errString := e.e.Error()
	return errString
}

func (e Error) GRPCStatus() *status.Status {
	return status.New(codes.Code(e.e.Code()), e.e.Message())
}

func CodeOf(err error) connect.Code {
	if code := connect.CodeOf(err); code != connect.CodeUnknown {
		return code
	}

	if code := status.Code(err); code != codes.Unknown {
		return connect.Code(code)
	}

	return connect.CodeUnknown
}

func MessageOf(err error) string {
	if err == nil {
		return ""
	}

	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce.Message()
	}

	if s, ok := status.FromError(err); ok {
		return s.Message()
	}

	return err.Error()
}

// IsCanceled indicates the operation was canceled (typically by the caller).
//
// The gRPC framework will generate this error code when cancellation
// is requested.
func IsCanceled(err error) bool {
	return CodeOf(err) == connect.CodeCanceled
}

// IsUnknown error. An example of where this error may be returned is
// if a Status value received from another address space belongs to
// an error-space that is not known in this address space. Also
// errors raised by APIs that do not return enough error information
// may be converted to this error.
//
// The gRPC framework will generate this error code in the above two
// mentioned cases.
func IsUnknown(err error) bool {
	return CodeOf(err) == connect.CodeUnknown
}

// IsInvalidArgument indicates client specified an invalid argument.
// Note that this differs from FailedPrecondition. It indicates arguments
// that are problematic regardless of the state of the system
// (e.g., a malformed file name).
//
// This error code will not be generated by the gRPC framework.
func IsInvalidArgument(err error) bool {
	return CodeOf(err) == connect.CodeInvalidArgument
}

// IsDeadlineExceeded means operation expired before completion.
// For operations that change the state of the system, this error may be
// returned even if the operation has completed successfully. For
// example, a successful response from a server could have been delayed
// long enough for the deadline to expire.
//
// The gRPC framework will generate this error code when the deadline is
// exceeded.
func IsDeadlineExceeded(err error) bool {
	return CodeOf(err) == connect.CodeDeadlineExceeded
}

// IsNotFound means some requested entity (e.g., file or directory) was
// not found.
//
// This error code will not be generated by the gRPC framework.
func IsNotFound(err error) bool {
	return CodeOf(err) == connect.CodeNotFound
}

// IsAlreadyExists means an attempt to create an entity failed because one
// already exists.
//
// This error code will not be generated by the gRPC framework.
func IsAlreadyExists(err error) bool {
	return CodeOf(err) == connect.CodeAlreadyExists
}

// IsPermissionDenied indicates the caller does not have permission to
// execute the specified operation. It must not be used for rejections
// caused by exhausting some resource (use ResourceExhausted
// instead for those errors). It must not be
// used if the caller cannot be identified (use Unauthenticated
// instead for those errors).
//
// This error code will not be generated by the gRPC core framework,
// but expect authentication middleware to use it.
func IsPermissionDenied(err error) bool {
	return CodeOf(err) == connect.CodePermissionDenied
}

// IsResourceExhausted indicates some resource has been exhausted, perhaps
// a per-user quota, or perhaps the entire file system is out of space.
//
// This error code will be generated by the gRPC framework in
// out-of-memory and server overload situations, or when a message is
// larger than the configured maximum size.
func IsResourceExhausted(err error) bool {
	return CodeOf(err) == connect.CodeResourceExhausted
}

// IsFailedPrecondition indicates operation was rejected because the
// system is not in a state required for the operation's execution.
// For example, directory to be deleted may be non-empty, an rmdir
// operation is applied to a non-directory, etc.
//
// A litmus test that may help a service implementor in deciding
// between FailedPrecondition, Aborted, and Unavailable:
//
//	(a) Use Unavailable if the client can retry just the failing call.
//	(b) Use Aborted if the client should retry at a higher-level
//	    (e.g., restarting a read-modify-write sequence).
//	(c) Use FailedPrecondition if the client should not retry until
//	    the system state has been explicitly fixed. E.g., if an "rmdir"
//	    fails because the directory is non-empty, FailedPrecondition
//	    should be returned since the client should not retry unless
//	    they have first fixed up the directory by deleting files from it.
//	(d) Use FailedPrecondition if the client performs conditional
//	    REST Get/Update/Delete on a resource and the resource on the
//	    server does not match the condition. E.g., conflicting
//	    read-modify-write on the same resource.
//
// This error code will not be generated by the gRPC framework.
func IsFailedPrecondition(err error) bool {
	return CodeOf(err) == connect.CodeFailedPrecondition
}

// IsAborted indicates the operation was aborted, typically due to a
// concurrency issue like sequencer check failures, transaction aborts,
// etc.
//
// See litmus test above for deciding between FailedPrecondition,
// Aborted, and Unavailable.
//
// This error code will not be generated by the gRPC framework.
func IsAborted(err error) bool {
	return CodeOf(err) == connect.CodeAborted
}

// IsOutOfRange means operation was attempted past the valid range.
// E.g., seeking or reading past end of file.
//
// Unlike InvalidArgument, this error indicates a problem that may
// be fixed if the system state changes. For example, a 32-bit file
// system will generate InvalidArgument if asked to read at an
// offset that is not in the range [0,2^32-1], but it will generate
// OutOfRange if asked to read from an offset past the current
// file size.
//
// There is a fair bit of overlap between FailedPrecondition and
// OutOfRange. We recommend using OutOfRange (the more specific
// error) when it applies so that callers who are iterating through
// a space can easily look for an OutOfRange error to detect when
// they are done.
//
// This error code will not be generated by the gRPC framework.
func IsOutOfRange(err error) bool {
	return CodeOf(err) == connect.CodeOutOfRange
}

// IsUnimplemented indicates operation is not implemented or not
// supported/enabled in this service.
//
// This error code will be generated by the gRPC framework. Most
// commonly, you will see this error code when a method implementation
// is missing on the server. It can also be generated for unknown
// compression algorithms or a disagreement as to whether an RPC should
// be streaming.
func IsUnimplemented(err error) bool {
	return CodeOf(err) == connect.CodeUnimplemented
}

// IsInternal errors. Means some invariants expected by underlying
// system has been broken. If you see one of these errors,
// something is very broken.
//
// This error code will be generated by the gRPC framework in several
// internal error conditions.
func IsInternal(err error) bool {
	return CodeOf(err) == connect.CodeInternal
}

// IsUnavailable indicates the service is currently unavailable.
// This is a most likely a transient condition and may be corrected
// by retrying with a backoff. Note that it is not always safe to retry
// non-idempotent operations.
//
// See litmus test above for deciding between FailedPrecondition,
// Aborted, and Unavailable.
//
// This error code will be generated by the gRPC framework during
// abrupt shutdown of a server process or network connection.
func IsUnavailable(err error) bool {
	return CodeOf(err) == connect.CodeUnavailable
}

// IsDataLoss indicates unrecoverable data loss or corruption.
//
// This error code will not be generated by the gRPC framework.
func IsDataLoss(err error) bool {
	return CodeOf(err) == connect.CodeDataLoss
}

// IsUnauthenticated indicates the request does not have valid
// authentication credentials for the operation.
//
// The gRPC framework will generate this error code when the
// authentication metadata is invalid or a Credentials callback fails,
// but also expect authentication middleware to generate it.
func IsUnauthenticated(err error) bool {
	return CodeOf(err) == connect.CodeUnauthenticated
}

// CanceledErrorf creates a new error that indicates the operation was canceled (typically by the caller).
//
// The gRPC framework will generate this error code when cancellation
// is requested.
func CanceledErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeCanceled, fmt.Errorf(format, a...))
}

// UnknownErrorf creates a new error that error. An example of where this error may be returned is
// if a Status value received from another address space belongs to
// an error-space that is not known in this address space. Also
// errors raised by APIs that do not return enough error information
// may be converted to this error.
//
// The gRPC framework will generate this error code in the above two
// mentioned cases.
func UnknownErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeUnknown, fmt.Errorf(format, a...))
}

// InvalidArgumentErrorf creates a new error that indicates client specified an invalid argument.
// Note that this differs from FailedPrecondition. It indicates arguments
// that are problematic regardless of the state of the system
// (e.g., a malformed file name).
//
// This error code will not be generated by the gRPC framework.
func InvalidArgumentErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeInvalidArgument, fmt.Errorf(format, a...))
}

// DeadlineExceededErrorf creates a new error that means operation expired before completion.
// For operations that change the state of the system, this error may be
// returned even if the operation has completed successfully. For
// example, a successful response from a server could have been delayed
// long enough for the deadline to expire.
//
// The gRPC framework will generate this error code when the deadline is
// exceeded.
func DeadlineExceededErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeDeadlineExceeded, fmt.Errorf(format, a...))
}

// NotFoundErrorf creates a new error that means some requested entity (e.g., file or directory) was
// not found.
//
// This error code will not be generated by the gRPC framework.
func NotFoundErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeNotFound, fmt.Errorf(format, a...))
}

// AlreadyExistsErrorf creates a new error that means an attempt to create an entity failed because one
// already exists.
//
// This error code will not be generated by the gRPC framework.
func AlreadyExistsErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeAlreadyExists, fmt.Errorf(format, a...))
}

// PermissionDeniedErrorf creates a new error that indicates the caller does not have permission to
// execute the specified operation. It must not be used for rejections
// caused by exhausting some resource (use ResourceExhausted
// instead for those errors). It must not be
// used if the caller cannot be identified (use Unauthenticated
// instead for those errors).
//
// This error code will not be generated by the gRPC core framework,
// but expect authentication middleware to use it.
func PermissionDeniedErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodePermissionDenied, fmt.Errorf(format, a...))
}

// ResourceExhaustedErrorf creates a new error that indicates some resource has been exhausted, perhaps
// a per-user quota, or perhaps the entire file system is out of space.
//
// This error code will be generated by the gRPC framework in
// out-of-memory and server overload situations, or when a message is
// larger than the configured maximum size.
func ResourceExhaustedErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeResourceExhausted, fmt.Errorf(format, a...))
}

// FailedPreconditionErrorf creates a new error that indicates operation was rejected because the
// system is not in a state required for the operation's execution.
// For example, directory to be deleted may be non-empty, an rmdir
// operation is applied to a non-directory, etc.
//
// A litmus test that may help a service implementor in deciding
// between FailedPrecondition, Aborted, and Unavailable:
//
//	(a) Use Unavailable if the client can retry just the failing call.
//	(b) Use Aborted if the client should retry at a higher-level
//	    (e.g., restarting a read-modify-write sequence).
//	(c) Use FailedPrecondition if the client should not retry until
//	    the system state has been explicitly fixed. E.g., if an "rmdir"
//	    fails because the directory is non-empty, FailedPrecondition
//	    should be returned since the client should not retry unless
//	    they have first fixed up the directory by deleting files from it.
//	(d) Use FailedPrecondition if the client performs conditional
//	    REST Get/Update/Delete on a resource and the resource on the
//	    server does not match the condition. E.g., conflicting
//	    read-modify-write on the same resource.
//
// This error code will not be generated by the gRPC framework.
func FailedPreconditionErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeFailedPrecondition, fmt.Errorf(format, a...))
}

// AbortedErrorf creates a new error that indicates the operation was aborted, typically due to a
// concurrency issue like sequencer check failures, transaction aborts,
// etc.
//
// See litmus test above for deciding between FailedPrecondition,
// Aborted, and Unavailable.
//
// This error code will not be generated by the gRPC framework.
func AbortedErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeAborted, fmt.Errorf(format, a...))
}

// OutOfRangeErrorf creates a new error that means operation was attempted past the valid range.
// E.g., seeking or reading past end of file.
//
// Unlike InvalidArgument, this error indicates a problem that may
// be fixed if the system state changes. For example, a 32-bit file
// system will generate InvalidArgument if asked to read at an
// offset that is not in the range [0,2^32-1], but it will generate
// OutOfRange if asked to read from an offset past the current
// file size.
//
// There is a fair bit of overlap between FailedPrecondition and
// OutOfRange. We recommend using OutOfRange (the more specific
// error) when it applies so that callers who are iterating through
// a space can easily look for an OutOfRange error to detect when
// they are done.
//
// This error code will not be generated by the gRPC framework.
func OutOfRangeErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeOutOfRange, fmt.Errorf(format, a...))
}

// UnimplementedErrorf creates a new error that indicates operation is not implemented or not
// supported/enabled in this service.
//
// This error code will be generated by the gRPC framework. Most
// commonly, you will see this error code when a method implementation
// is missing on the server. It can also be generated for unknown
// compression algorithms or a disagreement as to whether an RPC should
// be streaming.
func UnimplementedErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeUnimplemented, fmt.Errorf(format, a...))
}

// InternalErrorf creates a new error that errors. Means some invariants expected by underlying
// system has been broken. If you see one of these errors,
// something is very broken.
//
// This error code will be generated by the gRPC framework in several
// internal error conditions.
func InternalErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeInternal, fmt.Errorf(format, a...))
}

// UnavailableErrorf creates a new error that indicates the service is currently unavailable.
// This is a most likely a transient condition and may be corrected
// by retrying with a backoff. Note that it is not always safe to retry
// non-idempotent operations.
//
// See litmus test above for deciding between FailedPrecondition,
// Aborted, and Unavailable.
//
// This error code will be generated by the gRPC framework during
// abrupt shutdown of a server process or network connection.
func UnavailableErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeUnavailable, fmt.Errorf(format, a...))
}

// DataLossErrorf creates a new error that indicates unrecoverable data loss or corruption.
//
// This error code will not be generated by the gRPC framework.
func DataLossErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeDataLoss, fmt.Errorf(format, a...))
}

// UnauthenticatedErrorf creates a new error that indicates the request does not have valid
// authentication credentials for the operation.
//
// The gRPC framework will generate this error code when the
// authentication metadata is invalid or a Credentials callback fails,
// but also expect authentication middleware to generate it.
func UnauthenticatedErrorf(format string, a ...interface{}) error {
	return NewError(connect.CodeUnauthenticated, fmt.Errorf(format, a...))
}

func NewError(c connect.Code, underlying error) error {
	e := connect.NewError(c, underlying)

	return Error{
		e: *e,
	}
}

func ToHTTP(err error) int {
	switch connect.CodeOf(err) {
	case connect.CodeCanceled:
		return http.StatusRequestTimeout
	case connect.CodeUnknown:
		return http.StatusInternalServerError
	case connect.CodeInvalidArgument:
		return http.StatusBadRequest
	case connect.CodeDeadlineExceeded:
		return http.StatusRequestTimeout
	case connect.CodeNotFound:
		return http.StatusNotFound
	case connect.CodeAlreadyExists:
		return http.StatusConflict
	case connect.CodePermissionDenied:
		return http.StatusForbidden
	case connect.CodeResourceExhausted:
		return http.StatusTooManyRequests
	case connect.CodeFailedPrecondition:
		return http.StatusPreconditionFailed
	case connect.CodeAborted:
		return http.StatusConflict
	case connect.CodeOutOfRange:
		return http.StatusBadRequest
	case connect.CodeUnimplemented:
		return http.StatusNotFound
	case connect.CodeInternal:
		return http.StatusInternalServerError
	case connect.CodeUnavailable:
		return http.StatusServiceUnavailable
	case connect.CodeDataLoss:
		return http.StatusInternalServerError
	case connect.CodeUnauthenticated:
		return http.StatusUnauthorized

	default:
		return http.StatusInternalServerError
	}
}

func FromHTTP(status int, msg string) error {
	switch status {
	case http.StatusBadRequest:
		return InvalidArgumentErrorf(msg)
	case http.StatusUnauthorized:
		return UnauthenticatedErrorf(msg)
	case http.StatusForbidden:
		return PermissionDeniedErrorf(msg)
	case http.StatusNotFound:
		return UnimplementedErrorf(msg)
	case http.StatusConflict:
		return AbortedErrorf(msg)
	default:
		return InternalErrorf(msg)
	}
}

func New(s string) error {
	return errors.New(s)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}
