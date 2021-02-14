package hook

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/dwellersclub/contigus/utils"
)

//NewRouter Create a new router
func NewRouter(logger *logrus.Logger, service Service, hookServerConfig utils.HookServerConfig, authFilter utils.AuthFilterFunc) *Router {
	return &Router{
		BaseRouter:       utils.NewBaseRouter(logger, authFilter),
		service:          service,
		hookServerConfig: hookServerConfig,
	}
}

//Router Web router
type Router struct {
	utils.BaseRouter
	service          Service
	hookServerConfig utils.HookServerConfig
}

//Build creates default routes for our API
func (rt *Router) Build() *mux.Router {
	urls := []*utils.URLHandler{
		utils.NewURLHandler(fmt.Sprintf("%s/{ID}", rt.hookServerConfig.Context), http.HandlerFunc(rt.handleHook), false, []string{"POST", "PUT", "GET"}, false),
	}
	return rt.InitRouter(urls)
}
