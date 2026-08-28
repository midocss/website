package http

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/midocss/website/internal/auth"
	"github.com/midocss/website/internal/config"
	"github.com/midocss/website/internal/rbac"
	"github.com/midocss/website/internal/transport/http/handler"
	"github.com/midocss/website/internal/transport/http/middleware"
	"github.com/midocss/website/internal/transport/http/response"
	"github.com/midocss/website/pkg/apperr"
)

// Handlers groups every HTTP handler mounted by the router.
type Handlers struct {
	Health *handler.HealthHandler
	Auth   *handler.AuthHandler
	Users  *handler.UserHandler
}

// Dependencies are the shared services the routes need.
type Dependencies struct {
	Config     *config.Config
	Logger     *slog.Logger
	Tokens     *auth.TokenManager
	Authorizer rbac.Authorizer
	Handlers   Handlers
}

// NewRouter wires the middleware chain and every route of the monolith.
func NewRouter(deps Dependencies) *gin.Engine {
	if deps.Config.IsProduction() || !deps.Config.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	if len(deps.Config.HTTP.TrustedProxies) > 0 {
		_ = router.SetTrustedProxies(deps.Config.HTTP.TrustedProxies)
	} else {
		_ = router.SetTrustedProxies(nil)
	}

	router.Use(
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.Logger(deps.Logger),
		middleware.SecurityHeaders(),
		middleware.CORS(deps.Config.HTTP.AllowedOrigins),
	)

	router.NoRoute(func(c *gin.Context) {
		response.Fail(c, apperr.NotFound("the requested endpoint does not exist"))
	})
	router.NoMethod(func(c *gin.Context) {
		response.Fail(c, apperr.New(apperr.CodeBadRequest, "method not allowed"))
	})

	router.GET("/health/live", deps.Handlers.Health.Live)
	router.GET("/health/ready", deps.Handlers.Health.Ready)

	// Login and registration are the prime credential-stuffing targets, so they
	// get a much tighter budget than the rest of the API.
	sensitive := middleware.RateLimit(middleware.RateLimitConfig{
		RequestsPerMinute: 10,
		Burst:             5,
		TTL:               15 * time.Minute,
	})
	standard := middleware.RateLimit(middleware.RateLimitConfig{
		RequestsPerMinute: 120,
		Burst:             60,
		TTL:               10 * time.Minute,
	})

	authenticated := middleware.Authenticate(deps.Tokens)

	v1 := router.Group("/api/v1")
	v1.Use(standard)
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", sensitive, deps.Handlers.Auth.Register)
			authGroup.POST("/login", sensitive, deps.Handlers.Auth.Login)
			authGroup.POST("/refresh", sensitive, deps.Handlers.Auth.Refresh)
			authGroup.POST("/logout", deps.Handlers.Auth.Logout)

			me := authGroup.Group("", authenticated)
			me.GET("/me", deps.Handlers.Auth.Me)
			me.POST("/logout-all", deps.Handlers.Auth.LogoutAll)
		}

		admin := v1.Group("/admin", authenticated)
		{
			users := admin.Group("/users")
			users.GET("", middleware.RequirePermissions(deps.Authorizer, rbac.UsersView), deps.Handlers.Users.List)
			users.GET("/:id", middleware.RequirePermissions(deps.Authorizer, rbac.UsersView), deps.Handlers.Users.Get)
			users.POST("", middleware.RequirePermissions(deps.Authorizer, rbac.UsersCreate), deps.Handlers.Users.Create)
			users.PATCH("/:id", middleware.RequirePermissions(deps.Authorizer, rbac.UsersUpdate), deps.Handlers.Users.Update)
			users.DELETE("/:id", middleware.RequirePermissions(deps.Authorizer, rbac.UsersDelete), deps.Handlers.Users.Delete)
			users.PUT("/:id/permissions", middleware.RequirePermissions(deps.Authorizer, rbac.RolesManage), deps.Handlers.Users.SetPermissions)

			admin.GET("/roles", middleware.RequirePermissions(deps.Authorizer, rbac.RolesView), deps.Handlers.Users.ListRoles)
			admin.GET("/permissions", middleware.RequirePermissions(deps.Authorizer, rbac.RolesView), deps.Handlers.Users.ListPermissions)
		}
	}

	return router
}

// NewServer builds the HTTP server with the configured timeouts.
func NewServer(cfg config.HTTP, router http.Handler) *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler:           router,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}
}
