package httpserver

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"strings"
)

type LoggingHandler interface {
	http.Handler
	GetLog() logr.Logger
}

type BaseHandler struct {
	Logger logr.Logger
}

// Each REST call must be concluded by one of these three functions

func (h *BaseHandler) ServeJSON(response http.ResponseWriter, data interface{}) {
	response.Header().Set("Content-Type", "application/json")
	if h.Logger.V(1).Enabled() {
		h.Logger.V(1).Info("<----- Emit JSON", "json", json2String(data))
	}
	response.WriteHeader(http.StatusOK)
	err := json.NewEncoder(response).Encode(data)
	if err != nil {
		panic(err)
	}
}

func (h *BaseHandler) HttpError(response http.ResponseWriter, message string, httpCode int) {
	if h.Logger.V(1).Enabled() {
		h.Logger.V(1).Info("<----- httpError", "message", message, "httpCode", httpCode)
	}
	http.Error(response, message, httpCode)
}

func (h *BaseHandler) HttpClose(response http.ResponseWriter, message string, httpCode int) {
	if h.Logger.V(1).Enabled() {
		h.Logger.V(1).Info("<----- httpClose", "message", message, "httpCode", httpCode)
	}
	if message != "" {
		response.Header().Set("Content-Type", "text/plain; charset=utf-8")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.WriteHeader(httpCode)
		_, _ = fmt.Fprintln(response, message)
	} else {
		response.WriteHeader(httpCode)
	}
}

func (h *BaseHandler) GetLog() logr.Logger {
	return h.Logger
}

func json2String(data interface{}) string {
	builder := &strings.Builder{}
	_ = json.NewEncoder(builder).Encode(data)
	return builder.String()
}

// ---------------------------------------------------

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
