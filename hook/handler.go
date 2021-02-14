package hook

import "net/http"

func (rt *Router) handleHook(w http.ResponseWriter, r *http.Request) {

	ID := rt.GetParam("ID", r)

	hook := rt.service.GetHook(ID)
	if hook == nil {
		rt.NotFound(w)
		return
	}

	hreq, err := rt.service.Read(r, hook)
	if rt.CheckError(r, w, err) {
		return
	}

	//dispatch request payload
	err = rt.service.Emit(hreq)
	if rt.CheckError(r, w, err) {
		return
	}
}
