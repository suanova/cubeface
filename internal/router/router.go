package router

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cubeface/cubeface/internal/common"
	"github.com/cubeface/cubeface/internal/config"
	"github.com/cubeface/cubeface/internal/controller"
	"github.com/cubeface/cubeface/internal/middleware"
	"github.com/cubeface/cubeface/internal/service"
	"github.com/cubeface/cubeface/internal/webui"
)

type Deps struct {
	Config       *config.Config
	Users        *controller.UserController
	Verification *controller.VerificationController
	OAuth        *controller.OAuthController
	Status       *controller.StatusController
	Models       *controller.ModelController
	Sessions     *service.SessionService
	Static       fs.FS
}

func New(d Deps) *gin.Engine {
	if d.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.Recovery(), middleware.SecurityHeaders())
	// An empty list disables proxy header trust, so ClientIP is RemoteAddr.
	_ = engine.SetTrustedProxies(d.Config.TrustedProxies)

	basePath := d.Config.BasePath
	api := engine.Group(basePath + "/api")
	api.Use(middleware.LimitBody(), middleware.RequireSameOrigin(d.Config.PublicOrigin))

	api.GET("/status", d.Status.Get)
	api.GET("/models", d.Models.List)
	api.GET("/models/:name", d.Models.Get)
	api.POST("/user/register", d.Users.Register)
	api.POST("/user/login", d.Users.Login)
	api.POST("/user/auth/logout", d.Users.Logout)
	api.POST("/verification/send", d.Verification.Send)
	api.GET("/oauth/:provider/start", d.OAuth.Start)
	api.GET("/oauth/:provider/callback", d.OAuth.Callback)

	protected := api.Group("")
	protected.Use(middleware.RequireAuth(d.Sessions))
	protected.GET("/user/self", d.Users.Self)

	engine.GET(basePath+"/healthz", d.Status.Health)

	static := webui.New(d.Static, basePath)
	apiPrefix := basePath + "/api"
	engine.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if requestPath == apiPrefix || strings.HasPrefix(requestPath, apiPrefix+"/") {
			common.Fail(c, common.NewAPIErrorStatus(http.StatusNotFound, "NOT_FOUND", "API endpoint not found"))
			return
		}
		static.ServeHTTP(c.Writer, c.Request)
	})

	return engine
}
