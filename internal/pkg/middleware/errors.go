package middleware

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errUnauthenticated   = status.Error(codes.Unauthenticated, "invalid credentials")
	errInternal          = status.Error(codes.Internal, "internal server error")
	errPermissionDenied  = status.Error(codes.PermissionDenied, "user does not have permission to access this endpoint")
	errDeadlineExceeded  = status.Error(codes.DeadlineExceeded, "request timed out")
	errResourceExhausted = status.Error(codes.ResourceExhausted, "too many requests")
)
