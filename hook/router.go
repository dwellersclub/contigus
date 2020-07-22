package hook

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/dwellersclub/contigus/utils"
)

//NewRouter Create a new router
func NewRouter(logger *logrus.Logger) *Router {
	return &Router{
		logger: logger,
	}
}

//Router Web router
type Router struct {
	AuthFilter utils.AuthFilterFunc
	logger     *logrus.Logger
}

//Build creates default routes for our API
func (rt *Router) Build() *mux.Router {
	fs := http.FileServer(http.Dir("./ui/dist"))
	urls := []*utils.URLHandler{
		utils.NewURLHandler("/", http.HandlerFunc(fs.ServeHTTP), false, []string{}, false),
	}
	return utils.InitRouter(urls, rt.AuthFilter, rt.logger)
}
