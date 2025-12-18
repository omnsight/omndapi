// jwt_middleware.go
package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ContextKey string

const (
	UserIDKey    ContextKey = "user_id"
	UserNameKey  ContextKey = "user_name"
	UserRolesKey ContextKey = "user_roles"
)

// IdentityInterceptor parses claims without verifying signature (Gateway trusted)
func GrpcGatewayIdentityInterceptor(clientID string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 1. Extract Token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing auth header")
		}
		tokenString := strings.TrimPrefix(values[0], "Bearer ")

		// 2. Parse Claims (Unverified because Gateway already verified it)
		parser := jwt.NewParser()
		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse token: %v", err)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, status.Error(codes.Internal, "invalid claims structure")
		}

		// 3. Extract User ID
		userName, _ := claims["preferred_username"].(string)

		// 4. Extract Roles for THIS specific Client
		var roles []string
		if resAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
			if clientMap, ok := resAccess[clientID].(map[string]interface{}); ok {
				if roleList, ok := clientMap["roles"].([]interface{}); ok {
					for _, r := range roleList {
						if rStr, ok := r.(string); ok {
							roles = append(roles, rStr)
						}
					}
				}
			}
		}

		// 5. Inject into Context
		ctx = context.WithValue(ctx, UserNameKey, userName)
		ctx = context.WithValue(ctx, UserRolesKey, roles)

		return handler(ctx, req)
	}
}

func GetUser(ctx context.Context) (string, []string, error) {
	userNameVal := ctx.Value(UserNameKey)
	rolesVal := ctx.Value(UserRolesKey)

	userName, ok := userNameVal.(string)
	if !ok || userName == "" {
		return "", nil, fmt.Errorf("user ID not found in context %v", userName)
	}

	roles, ok := rolesVal.([]string)
	if !ok {
		return "", nil, fmt.Errorf("user roles not found in context %v", roles)
	}

	return userName, roles, nil
}
