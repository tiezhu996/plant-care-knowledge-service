package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// UserKey is the gin context key for authenticated user claims.
const UserKey = "user"

// currentUserKey caches the user record loaded from the database during
// authentication so downstream handlers can reuse it without a second query.
const currentUserKey = "current_user"

// AuthRequired validates the JWT and injects claims into the context.
// After the token is verified it reloads the user's role from the database so
// that a role change (e.g. an admin being demoted) takes effect immediately,
// invalidating any still-valid token that still carries the old role.
func AuthRequired(cfg *config.Config, userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Fail(constants.CodeUnauthorized, constants.MsgUnauthorized))
			return
		}
		claims, err := util.ParseToken(strings.TrimPrefix(header, "Bearer "), cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Fail(constants.CodeUnauthorized, constants.MsgUnauthorized))
			return
		}
		// Refresh the role from the source of truth so stale tokens lose any
		// privileges that have since been revoked.
		if userRepo != nil {
			if u, ferr := userRepo.FindByID(claims.UserID); ferr == nil {
				claims.Role = u.Role
				c.Set(currentUserKey, u)
			} else if ferr == repository.ErrNotFound {
				// Account no longer exists: the token is no longer valid.
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Fail(constants.CodeUnauthorized, constants.MsgUnauthorized))
				return
			}
		}
		c.Set(UserKey, claims)
		c.Next()
	}
}

// GetUserID extracts the authenticated user id from the context.
func GetUserID(c *gin.Context) uint {
	v, ok := c.Get(UserKey)
	if !ok {
		return 0
	}
	claims, ok := v.(*util.Claims)
	if !ok {
		return 0
	}
	return claims.UserID
}

// GetUserRole extracts the authenticated user role from the context.
func GetUserRole(c *gin.Context) string {
	v, ok := c.Get(UserKey)
	if !ok {
		return ""
	}
	claims, ok := v.(*util.Claims)
	if !ok {
		return ""
	}
	return claims.Role
}

// CurrentUser returns the user record loaded from the database during
// authentication, or nil if none is cached.
func CurrentUser(c *gin.Context) *model.User {
	v, ok := c.Get(currentUserKey)
	if !ok {
		return nil
	}
	u, ok := v.(*model.User)
	if !ok {
		return nil
	}
	return u
}
