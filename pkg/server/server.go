package server

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/fatih/color"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/hiromaily/go-gin-wrapper/pkg/config"
	"github.com/hiromaily/go-gin-wrapper/pkg/files"
	"github.com/hiromaily/go-gin-wrapper/pkg/repository"
	"github.com/hiromaily/go-gin-wrapper/pkg/reverseproxy/types"
	"github.com/hiromaily/go-gin-wrapper/pkg/server/controller"
	"github.com/hiromaily/go-gin-wrapper/pkg/server/fcgi"
	sess "github.com/hiromaily/go-gin-wrapper/pkg/server/ginsession"
)

// Server interface
type Server interface {
	Start() (*gin.Engine, error)
	Close()
}

type server struct {
	gin          *gin.Engine
	port         int
	sessionStore sessions.Store
	middleware   Middlewarer
	controller   controller.Controller
	logger       *zap.Logger
	dbConn       *sql.DB
	userRepo     repository.UserRepository

	serverConf  *config.Server
	proxyConf   *config.Proxy
	apiConf     *config.API
	redisConf   *config.Redis
	developConf *config.Develop

	isTestMode bool
}

// NewServer returns Server interface
func NewServer(
	gin *gin.Engine,
	sessionStore sessions.Store,
	middleware Middlewarer,
	controller controller.Controller,
	logger *zap.Logger,
	dbConn *sql.DB,
	userRepo repository.UserRepository,
	conf *config.Root,
	isTestMode bool,
) Server {
	return &server{
		gin:          gin,
		port:         conf.Server.Port,
		sessionStore: sessionStore,
		middleware:   middleware,
		controller:   controller,
		logger:       logger,
		dbConn:       dbConn,
		userRepo:     userRepo,
		serverConf:   conf.Server,
		proxyConf:    conf.Proxy,
		apiConf:      conf.API,
		redisConf:    conf.Redis,
		developConf:  conf.Develop,
		isTestMode:   isTestMode,
	}
}

// Start starts gin server
func (s *server) Start() (*gin.Engine, error) {
	s.logger.Info("server Start()")

	s.setMiddleware()

	if err := s.loadTemplates(); err != nil {
		return nil, err
	}

	if err := s.loadStaticFiles(); err != nil {
		return nil, err
	}

	s.setRouter(s.gin)

	// set profiling for development use
	if s.developConf.ProfileEnable {
		ginpprof.Wrapper(s.gin)
	}

	// s.run() is not required if working for unittest
	if s.isTestMode {
		return s.gin, nil
	}

	err := s.run()
	return nil, err
}

// Close cleans up any middleware when shutdown server
func (s *server) Close() {
	s.logger.Info("server Close()")

	if s.dbConn != nil {
		s.dbConn.Close()
	}
}

func (s *server) setMiddleware() {
	s.logger.Info("server setMiddleware()")

	s.gin.Use(gin.Logger())

	s.gin.Use(s.middleware.GlobalRecover()) // It's called faster than [gin.Recovery()]

	// session
	sess.SetOption(s.sessionStore, s.serverConf.Session)
	s.gin.Use(sessions.Sessions(s.serverConf.Session.Name, s.sessionStore))

	s.gin.Use(s.middleware.FilterIP())

	s.gin.Use(s.middleware.CheckGinError())

	s.gin.Use(s.middleware.SetMetaData())

	s.gin.Use(s.middleware.UpdateUserSession())
}

func (s *server) projectRoot() (string, error) {
	projectPath := s.serverConf.Docs.Path
	// if projectPath includes ${GOPATH}, it should be replaced
	if strings.Contains(projectPath, "${GOPATH}") {
		gopath := os.Getenv("GOPATH")
		projectPath = strings.Replace(projectPath, "${GOPATH}", gopath, 1)
	}
	if f, err := os.Stat(projectPath); os.IsNotExist(err) || !f.IsDir() {
		// no such directory
		return "", err
	}
	return projectPath, nil
}

func (s *server) loadTemplates() error {
	s.logger.Info("server loadTemplates()")

	// it's impossible to call more than one. it was overwritten by final call.
	// r.LoadHTMLGlob("templates/*")
	// r.LoadHTMLGlob("templates/**/*")

	projectPath, err := s.projectRoot()
	if err != nil {
		return err
	}
	s.logger.Debug("loadTemplates()", zap.String("project_path", projectPath))

	ext := []string{"tmpl"}
	targetPath := []string{
		projectPath + "/web/templates/pages",
		projectPath + "/web/templates/components",
		projectPath + "/web/templates/inner_js",
	}
	var fileList []string
	for _, path := range targetPath {
		fl, err := files.GetFileList(path, ext)
		if err != nil {
			return err
		}
		fileList = append(fileList, fl...)
	}
	if len(fileList) == 0 {
		return errors.Errorf("file is not found in %s", projectPath)
	}

	tmpls := template.Must(template.New("").Funcs(getTempFunc()).ParseFiles(fileList...))
	s.gin.SetHTMLTemplate(tmpls)

	return nil
}

// template FuncMap
func getTempFunc() template.FuncMap {
	// type FuncMap map[string]interface{}
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"strAry": func(ary []string, i int) string {
			return ary[i]
		},
		"dateFmt": func(t time.Time) string {
			// fmt := "August 17, 2016 9:51 pm"
			// fmt := "2006-01-02 15:04:05"
			// fmt := "Mon Jan _2 15:04:05 2006"
			fmt := "Mon Jan _2 15:04:05"
			return t.Format(fmt)
		},
	}
	return funcMap
}

func (s *server) loadStaticFiles() error {
	projectPath, err := s.projectRoot()
	if err != nil {
		return err
	}
	s.logger.Info("server loadStaticFiles()", zap.String("project_path", projectPath))

	s.gin.Static("/statics", projectPath+"/web/statics")
	s.gin.Static("/assets", projectPath+"/web/statics/assets")
	s.gin.Static("/favicon.ico", projectPath+"/web/statics/favicon.ico")
	s.gin.Static("/swagger", projectPath+"/web/swagger/swagger-ui")

	return nil
}

func (s *server) run() error {
	s.logger.Info("server run()")
	addr := fmt.Sprintf(":%d", s.port)
	if s.proxyConf.Mode == types.NginxProxy {
		s.runFCGI(addr)
	}
	return s.runGin(addr)
}

func (s *server) runFCGI(addr string) error {
	// Proxy(Nginx) settings
	color.Red("[WARNING] server as fcgi mode.")
	s.logger.Info("running server as fcgi mode.")
	return fcgi.Run(s.gin, addr)
}

// how to shutdown with gin
// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/server.go
func (s *server) runGin(addr string) error {
	// s.gin.Run() would not return until error happens or detecting signal
	// return s.gin.Run(fmt.Sprintf(":%d", s.port))
	s.logger.Debug(fmt.Sprintf("Listening and serving HTTP on %s", addr))

	srv := &http.Server{
		Addr:    addr,
		Handler: s.gin,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("fail to call ListenAndServe()", zap.Error(err))
		}
	}()
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	s.logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
		s.Close()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error("fatal to call Shutdown():", zap.Error(err))
		return err
	}
	return nil
}
