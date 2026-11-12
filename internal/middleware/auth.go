package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/suanova/medrouter/internal/common"
	"github.com/suanova/medrouter/internal/model"
	"github.com/suanova/medrouter/internal/service"
)

const (
	ContextUserKey    = "medrouter.user"
	ContextSessionKey = "medrouter.session"
)

// RequireAuth resolves the opaque session cookie and injects the current user.
func RequireAuth(sessions *service.SessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessions.CookieName())
		if err != nil {
			common.AbortFail(c, common.NewAPIErrorStatus(http.StatusUnauthorized, "UNAUTHENTICATED", "Please sign in"))
			return
		}
		user, session, apiErr := sessions.Resolve(token)
		if apiErr != nil {
			common.AbortFail(c, apiErr)
			return
		}
		c.Set(ContextUserKey, user)
		c.Set(ContextSessionKey, session)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) *model.User {
	value, ok := c.Get(ContextUserKey)
	if !ok {
		return nil
	}
	user, _ := value.(*model.User)
	return user
}

func CurrentSession(c *gin.Context) *model.UserSession {
	value, ok := c.Get(ContextSessionKey)
	if !ok {
		return nil
	}
	session, _ := value.(*model.UserSession)
	return session
}
