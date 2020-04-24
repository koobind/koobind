package authserver

import (
	"fmt"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
)

type LoggedHandler interface {
	http.Handler
}

var httpLog = ctrl.Log.WithName("http")

func LogHttp(h LoggedHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if httpLog.V(1).Enabled() {
			httpLog.V(1).Info(fmt.Sprintf("--------- %s %s (from %s)", r.Method, r.RequestURI, r.RemoteAddr))
			if httpLog.V(2).Enabled() {
				for hdr := range r.Header {
					httpLog.V(2).Info(fmt.Sprintf("Header:%s - > %v", hdr, r.Header[hdr]))
				}

			}

		}
		h.ServeHTTP(w, r)
	})
}
