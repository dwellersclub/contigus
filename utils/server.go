package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go/types"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	logging "github.com/dwellersclub/contigus/log"
)

//ServerConfig Config for our http server
type ServerConfig struct {
	Port           int    `env:"PORT"`
	Bind           string `env:"BIND" envDefault:"0.0.0.0"`
	ServiceAccount string `env:"SERVICE_ACCOUNT_PATH"`
	AllowedOrigins string `env:"CORS_ALLOWED_ORIGIN"`
	WebHost        string `env:"WEB_HOST"`
	SSLCert        string `env:"SSL_CERT"`
	SSLCA          string `env:"SSL_CA"`
	SSLKey         string `env:"SSL_KEY"`

	DBUsername string `env:"DB_USERNAME" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"eSbpCQ8pS5x33bi"`
	DBURL      string `env:"DB_URL"  envDefault:"postgres|postgresql://[username]:[password]@localhost/contigus?connect_timeout=10&application_name=[appname]&sslmode=disable"`
}

//VersionInfo Version details
type VersionInfo struct {
	Version    string
	Buildstamp string
	Githash    string
}

//StartServer start http server
func StartServer(serverConfig ServerConfig, appName string, module string, router *mux.Router, version VersionInfo) {

	var log = logging.GetLogger()
	var fullAppName = fmt.Sprintf(`%s/%s`, appName, module)

	router.Handle("/metrics", promhttp.Handler())

	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, fmt.Sprintf(`{"name": "%s","alive": true}`, fullAppName))
	})

	router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, fmt.Sprintf(`{"name": "%s","version": "%s","buildstamp": "%s","githash": "%s"}`, fullAppName,
			version.Version, version.Buildstamp, version.Githash))
	})

	sslEnabled := len(serverConfig.SSLKey) > 0

	log.WithFields(logrus.Fields{
		"app":         strings.Title(module),
		"url":         fmt.Sprintf("http://%s:%d", serverConfig.Bind, serverConfig.Port),
		"start":       time.Now().Format("Mon Jan _2 15:04:05 2006"),
		"version":     version.Version,
		"build_stamp": version.Buildstamp,
		"git_hash":    version.Githash,
		"allow_hosts": serverConfig.AllowedOrigins,
		"ssl":         fmt.Sprintf("%t", sslEnabled),
	}).Info("started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	options := []handlers.CORSOption{}

	if len(serverConfig.AllowedOrigins) > 0 {
		options = []handlers.CORSOption{
			handlers.AllowCredentials(),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", "X-CLIENT-ID", "X-Request-ID"}),
			handlers.AllowedOrigins(strings.Split(serverConfig.AllowedOrigins, ",")),
			handlers.MaxAge(600),
		}
	}

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(serverConfig.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handlers.CORS(options...)(router),
	}

	if sslEnabled {
		tlsConfig, err := getTLSConfig(true, serverConfig.SSLCert, serverConfig.SSLKey, serverConfig.SSLCA)
		if err != nil {
			log.Fatalf("%s server can't use certificates %s", appName, err)
			return
		}

		srv.TLSConfig = tlsConfig
	}

	go func() {
		if sslEnabled {
			if err := srv.ListenAndServeTLS(serverConfig.SSLCert, serverConfig.SSLKey); err != nil {
				log.Fatalf("%s server can't start %s", appName, err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil {
				log.Errorf("%s server can't start %s", appName, err)
			}
		}
	}()

	<-stop

	log.Info("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	log.Info("Server gracefully stopped")

}

func getTLSConfig(verify bool, cert, key, ca string) (*tls.Config, error) {
	var config tls.Config
	config.InsecureSkipVerify = true

	if verify {
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		certPool.AppendCertsFromPEM(file)
		config.RootCAs = certPool
		config.InsecureSkipVerify = false
	}

	_, errCert := os.Stat(cert)
	_, errKey := os.Stat(key)
	if errCert == nil || errKey == nil {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load X509 key pair: %v. Key encrpyted", err)
		}
		config.Certificates = []tls.Certificate{tlsCert}
	}
	config.MinVersion = tls.VersionTLS10

	return &config, nil
}

func logRequest(handler http.Handler, logger *logrus.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

//DeleteMethod DELETE Method
var DeleteMethod = []string{"DELETE"}

//PostMethod POST Method
var PostMethod = []string{"POST"}

//GetMethod GET Method
var GetMethod = []string{"GET"}

//PutMethod PUT Method
var PutMethod = []string{"PUT"}

//PostGetMethod POST,GET Method
var PostGetMethod = []string{"GET", "POST"}

//AuthFilterFunc filter to validate current user
type AuthFilterFunc func(http.HandlerFunc, *logrus.Logger, bool, ...string) http.HandlerFunc

//URLHandler URL definition for a handler
type URLHandler struct {
	path    string
	handler http.HandlerFunc
	secure  bool
	gzip    bool
	methods []string
	roles   []string
	T       types.Type
}

//NewURLHandler Create a new Handler
func NewURLHandler(path string, handler http.HandlerFunc, secure bool, methods []string, gzip bool, roles ...string) *URLHandler {
	return &URLHandler{path: path, handler: handler, secure: secure, methods: methods, gzip: gzip, roles: roles}
}

//InitRouter Initialise a router base on URLS configs
func InitRouter(urls []*URLHandler, authFilter AuthFilterFunc, rawLogger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()
	for i := 0; i < len(urls); i++ {
		urlConfig := urls[i]
		handler := urlConfig.handler
		if authFilter != nil {
			handler = authFilter(handler, rawLogger, urlConfig.secure, urlConfig.roles...)
		}

		logger := rawLogger.WithFields(logrus.Fields{
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

func applyMetrics(handler http.Handler, name string) http.Handler {
	reg := prometheus.DefaultRegisterer

	inFlightGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "in_flight_requests",
		Help:        "A gauge of requests currently being served by the wrapped handler.",
		ConstLabels: prometheus.Labels{"handler": name},
	})

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "api_requests_total",
			Help:        "A counter for requests to the wrapped handler.",
			ConstLabels: prometheus.Labels{"handler": name},
		},
		[]string{"code"},
	)

	histVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "response_duration_seconds",
			Help:        "A histogram of request latencies.",
			Buckets:     prometheus.DefBuckets,
			ConstLabels: prometheus.Labels{"handler": name},
		},
		[]string{"code"},
	)

	reg.Register(inFlightGauge)
	reg.Register(counter)
	reg.Register(histVec)

	chainedHandler := promhttp.InstrumentHandlerDuration(histVec, handler)
	chainedHandler = promhttp.InstrumentHandlerCounter(counter, chainedHandler)
	return promhttp.InstrumentHandlerInFlight(inFlightGauge, chainedHandler)
}
