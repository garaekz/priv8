package auth

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/internal/errors"
	"github.com/garaekz/priv8/pkg/log"
	"golang.org/x/oauth2"
)

// Service encapsulates the authentication logic.
type Service interface {
	// authenticate authenticates a user using username and password.
	// It returns a JWT token if authentication succeeds. Otherwise, an error is returned.
	Login(ctx context.Context, username, password string) (string, error)
	SocialAuth(ctx context.Context, provider, code string) (string, error)
}

// Identity represents an authenticated user identity.
type Identity interface {
	// GetID returns the user ID.
	GetID() string
	// GetName returns the user name.
	GetName() string
}

type service struct {
	signingKey      string
	tokenExpiration int
	logger          log.Logger
	authProviders   map[string]*oauth2.Config
}

// NewService creates a new authentication service.
func NewService(signingKey string, tokenExpiration int, logger log.Logger, authProviders map[string]*oauth2.Config) Service {
	return service{signingKey, tokenExpiration, logger, authProviders}
}

// Login authenticates a user and generates a JWT token if authentication succeeds.
// Otherwise, an error is returned.
func (s service) Login(ctx context.Context, username, password string) (string, error) {
	if identity := s.authenticate(ctx, username, password); identity != nil {
		return s.generateJWT(identity)
	}
	return "", errors.Unauthorized("")
}

// SocialAuth authenticates a user using a social network (e.g., Facebook) and generates a JWT token if authentication succeeds.
// Otherwise, an error is returned.
func (s service) SocialAuth(ctx context.Context, provider, code string) (string, error) {
	config, ok := s.authProviders[provider]

	if !ok {
		s.logger.With(ctx).Errorf("Failed to find provider: %s", provider)
		return "", errors.BadRequest("Provider not supported")
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		s.logger.With(ctx).Errorf("failed to exchange code: %v", err)
		return "", errors.Unauthorized("")
	}

	return token.AccessToken, nil
}

// authenticate authenticates a user using username and password.
// If username and password are correct, an identity is returned. Otherwise, nil is returned.
func (s service) authenticate(ctx context.Context, username, password string) Identity {
	logger := s.logger.With(ctx, "user", username)

	// TODO: the following authentication logic is only for demo purpose
	if username == "demo" && password == "pass" {
		logger.Infof("authentication successful")
		return entity.User{ID: "100", Name: "demo"}
	}

	logger.Infof("authentication failed")
	return nil
}

// generateJWT generates a JWT that encodes an identity.
func (s service) generateJWT(identity Identity) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   identity.GetID(),
		"name": identity.GetName(),
		"exp":  time.Now().Add(time.Duration(s.tokenExpiration) * time.Hour).Unix(),
	}).SignedString([]byte(s.signingKey))
}
