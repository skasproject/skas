package skserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"path/filepath"
	"skas/sk-common/pkg/config"
	"skas/sk-common/pkg/skserver/certwatcher"
	"skas/sk-common/pkg/skserver/handlers"
	"skas/sk-common/proto/v1/proto"
)

type SkServer interface {
	Run(ctx context.Context) error
	Start(ctx context.Context) error
	AddHandler(meta *proto.RequestMeta, handler http.Handler)
	GetLog() logr.Logger
}

var _ SkServer = &skServer{}

type skServer struct {
	Name   string
	Log    logr.Logger
	Config *config.SkServerConfig
	Router *mux.Router
}

func (server *skServer) AddHandler(meta *proto.RequestMeta, handler http.Handler) {
	lh, ok := handler.(LoggingHandler)
	if ok {
		lh.SetLog(server.Log.WithName(fmt.Sprintf("%s handler", meta.Name)))
		if lh.GetLog().GetSink() == nil {
			panic(fmt.Sprintf("Handler '%s' does not implements correctly LoggingHandler interface", meta.Name))
		}
		lh.GetLog().Info(fmt.Sprintf("'%s' service ENABLED", meta.Name))
	} else {
		// All our handlers should implements LogginHandler interface
		panic(fmt.Sprintf("Handler '%s' does not implements LoggingHandler interface", meta.Name))
	}
	server.Router.Handle(meta.UrlPath, handler).Methods(meta.Method)
}

func (server *skServer) GetLog() logr.Logger {
	return server.Log
}

func New(name string, conf *config.SkServerConfig, log logr.Logger) SkServer {
	server := &skServer{
		Name:   name,
		Log:    log,
		Config: conf,
	}
	if *server.Config.Ssl {
		if server.Config.CertName == "" {
			server.Config.CertName = "tls.crt"
		}
		if server.Config.KeyName == "" {
			server.Config.KeyName = "tls.key"
		}
	}
	if server.Router == nil {
		server.Router = mux.NewRouter()
		server.Router.Use(LogHttp)
		server.Router.MethodNotAllowedHandler = &handlers.MethodNotAllowedHandler{
			Logger: server.Log,
		}
		server.Router.NotFoundHandler = &handlers.NotFoundHandler{
			Logger: server.Log,
		}
	}
	return server
}

// This function for the runnable package

func (server *skServer) Run(ctx context.Context) error {
	return server.Start(ctx)
}

// This function for the kubebuilder manager

func (server *skServer) Start(ctx context.Context) error {
	server.Log.Info("Starting skServer")

	bindAddr := fmt.Sprintf("%s:%d", server.Config.Interface, server.Config.Port)

	var listener net.Listener
	var err error
	if !*server.Config.Ssl {
		listener, err = net.Listen("tcp", bindAddr)
		if err != nil {
			return fmt.Errorf("%s: Error on net.Listen(): %w", server.Name, err)
		}
	} else {
		if server.Config.CertDir == "" {
			return fmt.Errorf("%s: CertDir is not defined while 'ssl'' is true", server.Name)
		}
		certPath := filepath.Join(server.Config.CertDir, server.Config.CertName)
		keyPath := filepath.Join(server.Config.CertDir, server.Config.KeyName)
		certWatcher, err := certwatcher.New(server.Name, certPath, keyPath, server.Log)
		if err != nil {
			return fmt.Errorf("%s: Error on certwatcher.New(): %w", server.Name, err)
		}
		go func() {
			if err := certWatcher.Start(ctx); err != nil {
				server.Log.Error(err, "certificate watcher error")
			}
		}()

		cfg := &tls.Config{
			NextProtos:     []string{"h2"},
			GetCertificate: certWatcher.GetCertificate,
		}

		listener, err = tls.Listen("tcp", bindAddr, cfg)
		if err != nil {
			return fmt.Errorf("%s: Error on tls.Listen(): %w", server.Name, err)
		}
	}

	server.Log.Info("Listening", "bindAddr", bindAddr, "ssl", server.Config.Ssl)

	srv := &http.Server{
		Handler: server.Router,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()
		server.Log.Info("shutting down server")

		// TODO: use a context with reasonable timeout
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout
			server.Log.Error(err, "error shutting down the HTTP server")
		}
		close(idleConnsClosed)
	}()

	err = srv.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: Error on srv.Serve(): %w", server.Name, err)
	}
	server.Log.Info("skServer shutdown")
	<-idleConnsClosed
	return nil
}
