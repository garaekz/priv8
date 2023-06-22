package auth

import (
	"github.com/garaekz/priv8/internal/errors"
	"github.com/garaekz/priv8/pkg/log"
	routing "github.com/go-ozzo/ozzo-routing/v2"
)

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(rg *routing.RouteGroup, service Service, logger log.Logger) {
	rg.Post("/login", login(service, logger))
	rg.Post("/oauth/<provider>", social(service, logger))
}

// social returns a handler that handles user login request via a social network (e.g., Facebook)
func social(service Service, logger log.Logger) routing.Handler {
	return func(c *routing.Context) error {
		provider := c.Param("provider")

		var req struct {
			Code string `form:"code"`
		}

		if err := c.Read(&req); err != nil {
			logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
			return errors.BadRequest("")
		}

		token, err := service.SocialAuth(c.Request.Context(), provider, req.Code)
		if err != nil {
			return err
		}

		return c.Write(struct {
			AccessToken string `json:"access_token"`
		}{token})
	}
}

// login returns a handler that handles user login request.
func login(service Service, logger log.Logger) routing.Handler {
	return func(c *routing.Context) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.Read(&req); err != nil {
			logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
			return errors.BadRequest("")
		}

		token, err := service.Login(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			return err
		}
		return c.Write(struct {
			Token string `json:"token"`
		}{token})
	}
}
