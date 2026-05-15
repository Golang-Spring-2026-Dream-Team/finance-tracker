package grpc

import (
	"context"
	"strings"

	"finance-tracker/pkg/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const grpcUserIDKey contextKey = "auth_user_id"
const grpcRoleKey contextKey = "auth_role"

// authInterceptor extracts the access token from the "authorization" metadata,
// parses it as JWT, and stores userID/role in the context.
type authInterceptor struct {
	jwtSecret string
}

func newAuthInterceptor(jwtSecret string) *authInterceptor {
	return &authInterceptor{jwtSecret: jwtSecret}
}

func (i *authInterceptor) UnaryInterceptor(ctx context.Context, req interface{},
	_ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	token, err := extractBearerToken(md)
	if err != nil {
		return nil, err
	}

	claims, err := auth.ParseAccessToken(i.jwtSecret, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	ctx = context.WithValue(ctx, grpcUserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, grpcRoleKey, claims.Role)

	return handler(ctx, req)
}

func extractBearerToken(md metadata.MD) (string, error) {
	authVals := md.Get("authorization")
	if len(authVals) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := authVals[0]
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", status.Error(codes.Unauthenticated, "empty bearer token")
	}

	return token, nil
}

// UserIDFromContext extracts the authenticated userID from gRPC context.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(grpcUserIDKey).(int64)
	return v, ok
}

// RoleFromContext extracts the authenticated user role from gRPC context.
func RoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(grpcRoleKey).(string)
	return v, ok
}
