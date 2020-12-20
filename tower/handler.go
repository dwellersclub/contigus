package tower

import (
	"net/http"
)

func (rt *Router) handleHook(w http.ResponseWriter, r *http.Request) {

	hook := rt.service.MatchURL(r.RequestURI)
	if hook == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hreq, err := rt.service.NewRequest(r, hook)
	if rt.checkError(r, w, err) {
		return
	}

	//dispatch request
	err = rt.service.EmitRequest(hreq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (rt *Router) isInstalled(w http.ResponseWriter, r *http.Request) {}
