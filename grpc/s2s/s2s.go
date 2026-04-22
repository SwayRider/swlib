// Package s2s provides utilities for service-to-service gRPC communication
// with automatic token refresh on authentication failures.
package s2s

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetTokenFN is a function type that returns a fresh authentication token.
// It's called when the initial call fails with an Unauthenticated error.
type GetTokenFN func() (string, error)

// Call executes a gRPC call with automatic token refresh on authentication failure.
// If the call returns an Unauthenticated error, it fetches a new token using
// newTokenFN and retries the call once.
//
// This is useful for service-to-service calls where tokens may expire.
//
// Example:
//
//	result, err := s2s.Call(
//	    func(token string) (*pb.Response, error) {
//	        return client.SomeMethod(ctx, req, withAuth(token))
//	    },
//	    currentToken,
//	    func() (string, error) {
//	        return authClient.GetServiceToken()
//	    },
//	)
func Call[T any](
	callFn func(token string) (T, error),
	token string,
	newTokenFN GetTokenFN,
) (res T, err error) {
	res, err = callFn(token)
	if err != nil {
		if status, ok := status.FromError(err); ok {
			if status.Code() == codes.Unauthenticated {
				token, err = newTokenFN()
				if err != nil {
					return
				}
				return callFn(token)
			}
		}
		return
	}
	return
}
