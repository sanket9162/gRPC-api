package interceptors

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthenticationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("AuthenticationInterceptor started")

	skipMethods := map[string]bool{
		"/main.ExecsService/Login":          true,
		"/main.ExecsService/ForgotPassword": true,
		"/main.ExecsService/ResetPassword":  true,
	}

	if skipMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata unavilable")
	}

	authHeader, ok := md["authorization"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token unavailable")
	}

	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	jwtSecret := os.Getenv("JWT_SECRET")

	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "Unauthorized Access")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unathorized Access")
	}

	if !parsedToken.Valid {
		return nil, status.Error(codes.Unauthenticated, "Unathorized Access")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unathorized Access")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthrized Access")
	}

	userId := claims["userId"].(string)
	username := claims["user"].(string)
	expiresAt := claims["exp"].(string)

	newCtx := context.WithValue(ctx, ContextKey("role"), role)
	newCtx = context.WithValue(newCtx, ContextKey("userId"), userId)
	newCtx = context.WithValue(newCtx, ContextKey("username"), username)
	newCtx = context.WithValue(newCtx, ContextKey("expiresAt"), expiresAt)

	return handler(newCtx, req)
}

type ContextKey string
