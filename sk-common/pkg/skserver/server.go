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
	"skas/sk-common/pkg/skserver/certwatcher"
	"skas/sk-common/pkg/skserver/handlers"
)

type ServerConfig struct {
	BindAddr string `yaml:"bindAddr"`
	Ssl      bool   `yaml:"ssl"`
	CertDir  string `yaml:"certDir"`  // CertDir is the directory that contains the server key and certificate.
	CertName string `yaml:"certName"` // CertName is the server certificate name. Defaults to tls.crt.
	KeyName  string `yaml:"keyName"`  // KeyName is the server key name. Defaults to tls.key.
}

type SkServer struct {
	Name string

	Log logr.Logger

	Config *ServerConfig

	Router *mux.Router
}

func (server *SkServer) Groom() {
	if server.Config.Ssl {
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
	return
}

func (server *SkServer) Run(ctx context.Context) error {
	return server.Start(ctx)
}

func (server *SkServer) Start(ctx context.Context) error {
	server.Log.Info("Starting SkServer")

	var listener net.Listener
	var err error
	if !server.Config.Ssl {
		listener, err = net.Listen("tcp", server.Config.BindAddr)
		if err != nil {
			return err
		}
	} else {
		if server.Config.CertDir == "" {
			return fmt.Errorf("CertDir is not defined while NoSsl is false")
		}
		certPath := filepath.Join(server.Config.CertDir, server.Config.CertName)
		keyPath := filepath.Join(server.Config.CertDir, server.Config.KeyName)
		certWatcher, err := certwatcher.New(server.Name, certPath, keyPath, server.Log)
		if err != nil {
			return err
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

		listener, err = tls.Listen("tcp", server.Config.BindAddr, cfg)
		if err != nil {
			return err
		}
	}

	server.Log.Info("Listening", "bindAddr", server.Config.BindAddr, "ssl", server.Config.Ssl)

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
		return err
	}
	server.Log.Info("SkServer shutdown")
	<-idleConnsClosed
	return nil
}
