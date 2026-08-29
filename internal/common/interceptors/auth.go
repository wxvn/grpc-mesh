package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthValidator interface {
	ValidateToken(ctx context.Context, accessToken string) (string, error)
}

var UserIDKey contextKey = "user_id"

func Auth(validator AuthValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(
				codes.Unauthenticated,
				"authorization metadata is missing",
			)
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(
				codes.Unauthenticated,
				"authorization token is missing",
			)
		}

		token := values[0]
		prefix := "Bearer "

		if !strings.HasPrefix(token, prefix) {
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid authorization format",
			)
		}

		token = strings.TrimPrefix(token, prefix)

		if token == "" {
			return nil, status.Error(
				codes.Unauthenticated,
				"authorization token is empty",
			)
		}

		userID, err := validator.ValidateToken(ctx, token)
		if err != nil {
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid access token",
			)
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)

		return handler(ctx, req)
	}
}

func UserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}

	return userID
}
