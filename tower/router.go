package tower

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/dwellersclub/contigus/hook"
	"github.com/dwellersclub/contigus/models"
	"github.com/dwellersclub/contigus/utils"
)

//NewRouter Create a new router
func NewRouter(logger *logrus.Logger, service hook.Service, config models.HookConfig) *Router {
	return &Router{
		logger:  logger,
		service: service,
		config:  config,
	}
}

//Router Web router
type Router struct {
	AuthFilter utils.AuthFilterFunc
	logger     *logrus.Logger
	service    hook.Service
	config     models.HookConfig
}

//Build creates default routes for our API
func (rt *Router) Build() *mux.Router {
	fs := http.FileServer(http.Dir("./ui/dist"))
	urls := []*utils.URLHandler{
		utils.NewURLHandler("/installed", http.HandlerFunc(rt.isInstalled), false, []string{}, false),
		utils.NewURLHandler(fmt.Sprintf("%s/{ID}", rt.config.URLContext), http.HandlerFunc(rt.handleHook), false, []string{"POST", "GET"}, false),
		utils.NewURLHandler("/", http.HandlerFunc(fs.ServeHTTP), false, []string{}, false),
	}
	return utils.InitRouter(urls, rt.AuthFilter, rt.logger)
}

//Encode encode model to JSON
func (rt *Router) Encode(w io.Writer, model interface{}) error {
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(model); err != nil {
		rt.logger.WithError(err).Error("Can't encode object")
		return err
	}
	return nil
}

//GetParam get a parameter from the url mapping
func (rt *Router) GetParam(key string, r *http.Request) string {
	vars := mux.Vars(r)
	value, exists := vars[key]
	if exists {
		return value
	}
	return ""
}

func (rt *Router) checkError(r *http.Request, w http.ResponseWriter, err error) bool {
	if err != nil {
		rt.logger.WithError(err).Errorf("can't complete request %s", r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return false
}
