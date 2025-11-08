package utils

import (
	"context"
	"errors"
)

type ContextKey string

func AuthorizeUser(ctx context.Context, allowRoles ...string) error {
	userRole, ok := ctx.Value(ContextKey("role")).(string)
	if !ok {
		return errors.New("user not authorized for access: role not found")
	}

	for _, allowedRoles := range allowRoles {
		if allowedRoles == userRole {
			return nil
		}
	}

	return errors.New("user not authorized for access")

}
