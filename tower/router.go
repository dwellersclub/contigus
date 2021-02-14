package tower

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/dwellersclub/contigus/hook"
	"github.com/dwellersclub/contigus/utils"
)

//NewRouter Create a new router
func NewRouter(logger *logrus.Logger, service hook.Service, authFilter utils.AuthFilterFunc) *Router {
	return &Router{
		BaseRouter: utils.NewBaseRouter(logger, authFilter),
		service:    service,
	}
}

//Router Web router
type Router struct {
	utils.BaseRouter
	service hook.Service
}

//Build creates default routes for our API
func (rt *Router) Build() *mux.Router {
	fs := http.FileServer(http.Dir("./ui/dist"))
	urls := []*utils.URLHandler{
		utils.NewURLHandler("/installed", http.HandlerFunc(rt.isInstalled), false, []string{}, false),
		utils.NewURLHandler("/", http.HandlerFunc(fs.ServeHTTP), false, []string{}, false),
	}
	return rt.InitRouter(urls)
}
