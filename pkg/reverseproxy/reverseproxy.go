package reverseproxy

import (
	"container/ring"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/hiromaily/go-gin-wrapper/pkg/config"
)

// Server interface
type Server interface {
	Start() error
}

// ----------------------------------------------------------------------------
// Server
// ----------------------------------------------------------------------------

type server struct {
	logger     *zap.Logger
	serverConf *config.Server
	proxyConf  *config.Proxy
}

// NewServer returns Server interface
func NewServer(
	logger *zap.Logger,
	conf *config.Root) Server {
	srv := server{
		logger:     logger,
		serverConf: conf.Server,
		proxyConf:  conf.Proxy,
	}
	return &srv
}

// Start starts reverse proxy server
func (s *server) Start() error {
	ports := s.proxyConf.Server.WebPort

	if len(ports) == 1 {
		s.singleReverseProxy()
	} else if len(ports) > 1 {
		s.multipleReverseProxy()
	}
	return errors.New("port setting is not found in config")
}

// Single Reverse Proxy
func (s *server) singleReverseProxy() {
	s.logger.Info("singleReverseProxy")
	// Web Server
	// webserverURL := "http://127.0.0.1:9990"
	serverURL, _ := url.Parse(
		getURL(s.serverConf.Scheme, s.serverConf.Host, s.serverConf.Port),
	)

	// proxy Server
	s.logger.Info("proxy is runnig ...", zap.Int("port", s.proxyConf.Server.Port))
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.proxyConf.Server.Port),
		Handler: httputil.NewSingleHostReverseProxy(serverURL),
	}
	server.ListenAndServe()
}

// Multiple Reverse Proxy
func (s *server) multipleReverseProxy() {
	ports := s.proxyConf.Server.WebPort
	s.logger.Info("multipleReverseProxy", zap.Int("server_num", len(ports)))
	// as precondition, increment port number by one.

	// web servers
	hostRing := ring.New(len(ports))
	for _, port := range ports {
		url, _ := url.Parse(getURL(s.serverConf.Scheme, s.serverConf.Host, port))
		hostRing.Value = url
		hostRing = hostRing.Next()
	}

	mutex := sync.Mutex{}
	// access server alternately
	director := func(request *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		request.URL.Scheme = s.serverConf.Scheme
		request.URL.Host = hostRing.Value.(*url.URL).Host
		hostRing = hostRing.Next()
	}

	// proxy Server
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.proxyConf.Server.Port),
		Handler: &httputil.ReverseProxy{Director: director},
	}
	server.ListenAndServe()
}

func getURL(scheme, host string, port int) string {
	return fmt.Sprintf("%s://%s:%d", scheme, host, port)
}
