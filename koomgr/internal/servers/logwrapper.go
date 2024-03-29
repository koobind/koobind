/*
  Copyright (C) 2020 Serge ALEXANDRE

  This file is part of koobind project

  koobind is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  koobind is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/

package servers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	ctrl "sigs.k8s.io/controller-runtime"
)

var httpLog = ctrl.Log.WithName("http")

func LogHttp(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if httpLog.V(1).Enabled() {
			httpLog.V(1).Info(fmt.Sprintf("--------- %s %s (from %s)", r.Method, r.RequestURI, r.RemoteAddr))
			if httpLog.V(2).Enabled() {
				dump, err := httputil.DumpRequest(r, true)
				if err != nil {
					http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
					return
				}
				httpLog.V(2).Info(fmt.Sprintf("%q", dump))
				//for hdr := range r.Header {
				//	httpLog.V(2).Info(fmt.Sprintf("Header:%s - > %v", hdr, r.Header[hdr]))
				//}
			}
		}
		h.ServeHTTP(w, r)
	})
}
