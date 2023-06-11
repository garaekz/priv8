package secret

import (
	"net/http"

	"github.com/garaekz/priv8/internal/errors"
	"github.com/garaekz/priv8/pkg/log"
	routing "github.com/go-ozzo/ozzo-routing/v2"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, authHandler routing.Handler, logger log.Logger) {
	res := resource{service, logger}

	r.Post("/secrets", res.create)
	r.Get("/secrets/<id>", res.get)
	r.Post("/secrets/<id>", res.read)

	r.Use(authHandler)

	// the following endpoints require a valid JWT
	// r.Get("/secrets", res.query)
	// r.Put("/secrets/<id>", res.update)
	// r.Delete("/secrets/<id>", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c *routing.Context) error {
	secret, err := r.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(secret)
}

func (r resource) read(c *routing.Context) error {
	var input ReadSecretRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	secret, err := r.service.Read(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		return err
	}

	return c.Write(secret)
}

func (r resource) create(c *routing.Context) error {
	var input CreateSecretRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	secret, err := r.service.Create(c.Request.Context(), input)
	if err != nil {
		return err
	}

	return c.WriteWithStatus(secret, http.StatusCreated)
}

func (r resource) delete(c *routing.Context) error {
	secret, err := r.service.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(secret)
}
