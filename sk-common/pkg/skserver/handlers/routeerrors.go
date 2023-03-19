package handlers

import (
	"github.com/go-logr/logr"
	"net/http"
)

var _ http.Handler = &NotFoundHandler{}
var _ http.Handler = &MethodNotAllowedHandler{}

type NotFoundHandler struct {
	Logger logr.Logger
}

func (h *NotFoundHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.Logger.V(1).Info("Url not found", "uri", request.RequestURI)
	http.Error(writer, "", http.StatusNotFound)
}

type MethodNotAllowedHandler struct {
	Logger logr.Logger
}

func (h MethodNotAllowedHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.Logger.V(1).Info("Method not allowes", "method", request.Method)
	http.Error(writer, "", http.StatusMethodNotAllowed)
}
