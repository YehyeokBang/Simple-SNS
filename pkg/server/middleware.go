package server

import (
	"context"
	"strings"

	"github.com/YehyeokBang/Simple-SNS/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(jwt *auth.JWT) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/v1.user.UserService/SignUp" ||
			info.FullMethod == "/v1.user.UserService/LogIn" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		authorization, ok := md["authorization"]
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		token := strings.TrimPrefix(authorization[0], "Bearer ")
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "access denied: invalid token")
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "access denied: invalid 'sub' field in token")
		}

		ctx = context.WithValue(ctx, auth.UserIDKey, userID)

		return handler(ctx, req)
	}
}
