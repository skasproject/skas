package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"skas/sk-common/proto/v1/proto"
	"strings"
)

type BaseHandler struct {
	Logger logr.Logger
}

// Each REST call must be concluded by one of these three functions

func (h *BaseHandler) ServeJSON(response http.ResponseWriter, payload proto.ResponsePayload) {
	response.Header().Set("Content-Type", "application/json")
	if h.Logger.V(1).Enabled() {
		h.Logger.V(1).Info("<----- Emit JSON", "json", json2String(payload))
	}
	response.WriteHeader(http.StatusOK)
	err := json.NewEncoder(response).Encode(payload)
	if err != nil {
		panic(err)
	}
}

func (h *BaseHandler) HttpSendError(response http.ResponseWriter, message string, httpCode int) {
	if h.Logger.V(1).Enabled() {
		h.Logger.V(1).Error(nil, "<----- httpError", "message", message, "httpCode", httpCode)
	} else {
		h.Logger.Error(nil, "!!! http error", "message", message, "httpCode", httpCode)
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

func (h *BaseHandler) SetLog(logger logr.Logger) {
	h.Logger = logger
}

func json2String(data interface{}) string {
	builder := &strings.Builder{}
	_ = json.NewEncoder(builder).Encode(data)
	return builder.String()
}

// ---------------------------------------------------
