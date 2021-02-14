package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	validator "gopkg.in/asaskevich/govalidator.v9"
)

var blank = []byte(`{}`)

//NewBaseRouter Create new router with basic features
func NewBaseRouter(logger *logrus.Logger, authFilter AuthFilterFunc) BaseRouter {

	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	var hashKey = []byte("09876543210987654321098765432156")
	var blockKey = []byte("09876543210987654321098765432156")
	var s = securecookie.New(hashKey, blockKey)

	return BaseRouter{logger: logger, authFilter: authFilter, Decoder: decoder, MaxReadSize: 1024000,
		securecookie: s}
}

//BaseRouter base router with helper methods
type BaseRouter struct {
	logger       *logrus.Logger
	authFilter   AuthFilterFunc
	MaxReadSize  int64
	Decoder      *schema.Decoder
	securecookie *securecookie.SecureCookie
	enforcer     *casbin.CachedEnforcer
}

//Encode encode an struct to json
func (rt *BaseRouter) Encode(w http.ResponseWriter, model interface{}) error {
	rt.SetHeaderJSONType(w)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(model); err != nil {
		rt.logger.WithError(err).Error("Can't encode object")
		return err
	}
	return nil
}

//SetBadRequest set 400 status code and json response
func (rt *BaseRouter) SetBadRequest(w http.ResponseWriter) {
	rt.SetHeaderJSONType(w)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{ "errMsg" : "error_bad_request" }`))
}

//Not found set not found status
func (rt *BaseRouter) NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

//SetInternalError set 500 status code and json response
func (rt *BaseRouter) SetInternalError(w http.ResponseWriter) {
	rt.SetHeaderJSONType(w)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{ "errMsg" : "error_server_error" }`))
}

//SetBlankResponse blank json response
func (rt *BaseRouter) SetBlankResponse(w http.ResponseWriter) {
	rt.SetHeaderJSONType(w)
	w.Write(blank)
}

//BindForm http form to a struct, can't be validated based on annotations, see https://github.com/asaskevich/govalidator
func (rt *BaseRouter) BindForm(w http.ResponseWriter, r *http.Request, model interface{}, validate bool) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = rt.Decoder.Decode(model, r.PostForm)
	if err != nil {
		rt.SetBadRequest(w)
		return err
	}

	if validate {
		var valid bool
		valid, err = validator.ValidateStruct(model)
		if !valid {
			rt.SetBadRequest(w)
			return err
		}
	}

	return nil
}

//BindURL bind url query param to a struct
func (rt *BaseRouter) BindURL(w http.ResponseWriter, r *http.Request, model interface{}, validate bool) error {
	err := rt.Decoder.Decode(model, r.URL.Query())
	if err != nil {
		rt.SetBadRequest(w)
		return err
	}

	if validate {
		var valid bool
		valid, err = validator.ValidateStruct(model)
		if !valid {
			rt.SetBadRequest(w)
			return err
		}
	}

	return nil
}

//GetParam get a parameter from the url mapping
func (rt *BaseRouter) GetParam(key string, r *http.Request) string {
	vars := mux.Vars(r)
	value, exists := vars[key]
	if exists {
		return value
	}
	return ""
}

//ReadRawBody Read complete raw request body
func (rt *BaseRouter) ReadRawBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, rt.MaxReadSize))
	if err != nil {
		rt.logger.WithError(err).Error("Can't read body")
		rt.SetBadRequest(w)
		return nil, err
	}

	return body, err
}

//SetSecureCookie set secured cookies
func (rt *BaseRouter) SetSecureCookie(w http.ResponseWriter, cookieName string, cookieValue string, maxAge int) {
	value := map[string]string{cookieName: cookieValue}
	if encoded, err := rt.securecookie.Encode(cookieName, value); err == nil {
		cookie := &http.Cookie{
			Name:   cookieName,
			Value:  encoded,
			Path:   "/",
			Secure: true,
			MaxAge: maxAge,
		}
		http.SetCookie(w, cookie)
	}
}

//GetSecureCookie get secured cookies
func (rt *BaseRouter) GetSecureCookie(r *http.Request, cookieName string) string {
	if cookie, err := r.Cookie(cookieName); err == nil {
		value := make(map[string]string)
		if err = rt.securecookie.Decode(cookieName, cookie.Value, &value); err == nil {
			return value[cookieName]
		}
	}
	return ""
}

// ReadBody bing body paylod to a struct
func (rt *BaseRouter) ReadBody(w http.ResponseWriter, r *http.Request, model interface{}, validate bool) error {
	defer r.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, rt.MaxReadSize))
	if err != nil {
		rt.logger.WithError(err).Error("Can't read body")
		rt.SetBadRequest(w)
		return err
	}

	if err := json.Unmarshal(body, model); err != nil {
		rt.logger.WithError(err).Error("Can't parse body")
		rt.SetBadRequest(w)
		return err
	}

	if validate {
		var valid bool
		valid, err = validator.ValidateStruct(model)
		if !valid {
			rt.logger.WithError(err).Error("Validation failure")
			rt.SetBadRequest(w)
			return err
		}
	}

	return nil
}

//SetHeaderJSONType set content type to 'application/json
func (rt *BaseRouter) SetHeaderJSONType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

//SetCache Set header cache
func (rt *BaseRouter) SetCache(w http.ResponseWriter, minutes int64) {
	now := time.Now()
	t := now.Add(time.Duration(minutes) * time.Minute)
	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", minutes*60))
	w.Header().Set("Expires", t.Format(time.RFC1123))
	w.Header().Set("Date", now.Format(time.RFC1123))
}

//CheckError check if an error is not nil, if not nil will log an error and set response with 500 status
func (rt *BaseRouter) CheckError(r *http.Request, w http.ResponseWriter, err error) bool {
	if err != nil {
		rt.logger.WithError(err).Errorf("can't complete request %s", r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return false
}

//InitRouter Initialise a router base on URLS configs
func (rt *BaseRouter) InitRouter(urls []*URLHandler) *mux.Router {
	router := mux.NewRouter()
	for i := 0; i < len(urls); i++ {
		urlConfig := urls[i]
		handler := urlConfig.handler

		if rt.authFilter != nil {
			handler = rt.authFilter(handler, rt.logger, urlConfig.secure, urlConfig.roles...)
		}

		logger := rt.logger.WithFields(logrus.Fields{
			"secured": urlConfig.secure,
			"methods": urlConfig.methods,
			"path":    urlConfig.path,
			"gzip":    urlConfig.gzip,
		})

		wrappedHandler := applyMetrics(logRequest(handler, logger), urlConfig.path)

		if urlConfig.gzip {
			wrappedHandler = gziphandler.GzipHandler(wrappedHandler)
		}

		if len(urlConfig.methods) > 0 {
			router.Handle(urlConfig.path, wrappedHandler).Methods(urlConfig.methods...)
		} else {
			router.PathPrefix(urlConfig.path).Handler(wrappedHandler)
		}

		logger.Debug("Path configured")
	}

	return router
}
