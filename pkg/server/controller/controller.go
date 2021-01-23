package controller

import (
	"go.uber.org/zap"

	"github.com/hiromaily/go-gin-wrapper/pkg/auth/jwts"
	"github.com/hiromaily/go-gin-wrapper/pkg/config"
	"github.com/hiromaily/go-gin-wrapper/pkg/repository"
)

// Controller  interface
type Controller interface {
	Accounter
	Adminer
	APIJWTer
	APIUser
	APILister
	Baser
	Chater
	Errorer
	Loginer
	OAuther
}

type controller struct {
	logger        *zap.Logger
	userRepo      repository.UserRepository
	jwter         jwts.JWTer
	apiHeaderConf *config.Header
	authConf      *config.Auth
}

// NewController returns Controller
func NewController(
	logger *zap.Logger,
	userRepo repository.UserRepository,
	jwter jwts.JWTer,
	apiHeaderConf *config.Header,
	auth *config.Auth) Controller {
	return &controller{
		logger:        logger,
		userRepo:      userRepo,
		jwter:         jwter,
		apiHeaderConf: apiHeaderConf,
		authConf:      auth,
	}
}
