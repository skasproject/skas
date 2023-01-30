package httpserver

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
)

type LoggingHandler interface {
	http.Handler
	GetLog() logr.Logger
}

func LogHttp(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lh, ok := h.(LoggingHandler)
		if ok {
			if lh.GetLog().V(1).Enabled() {
				if lh.GetLog().V(2).Enabled() {
					dump, err := httputil.DumpRequest(r, true)
					if err != nil {
						http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
						return
					}
					lh.GetLog().V(2).Info("-----> HTTP Request", "request", dump)
					//lh.GetLog().V(2).Info(fmt.Sprintf("%q", dump))
					//for hdr := range r.Header {
					//	httpLog.V(2).Info(fmt.Sprintf("Header:%s - > %v", hdr, r.Header[hdr]))
					//}
				} else {
					lh.GetLog().V(1).Info("-----> HTTP Request", "method", r.Method, "uri", r.RequestURI, "remote", r.RemoteAddr)
				}
			}
		} else {
			// We don't have logger. We hack just to indicate this case. (All our Handler should support LoggingHandler)
			logrusLog := logrus.New()
			logrusLog.SetLevel(logrus.DebugLevel)
			logrusLog.Log(logrus.WarnLevel, "An handler does not implements LoggingHandler interface")
		}
		h.ServeHTTP(w, r)
	})
}
