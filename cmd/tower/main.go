package main

import (
	"os"

	"github.com/caarlos0/env"
	_ "github.com/lib/pq"

	"github.com/dwellersclub/contigus/hook"
	logging "github.com/dwellersclub/contigus/log"
	"github.com/dwellersclub/contigus/models"
	"github.com/dwellersclub/contigus/tower"
	"github.com/dwellersclub/contigus/utils"
)

var log = logging.GetLogger()
var (
	version    string
	buildstamp string
	githash    string
)

func main() {
	var appName = "tower"

	config := utils.ServerConfig{Port: 8081}

	err := env.Parse(&config)

	if err != nil {
		log.WithError(err).Errorf("Can't start %s", appName)
		os.Exit(1)
		return
	}

	dbURL := utils.GetDBUrl(config.DBURL, config.DBUsername, config.DBPassword, appName)

	//TODO: make it configurable
	hookConfig := models.HookConfig{
		URLContext: "/hooks",
	}

	encryptor := hook.NewEncryptor()
	metrics := hook.NewHookMetrics()

	service := hook.NewService(metrics, encryptor, "", nil)

	router := tower.NewRouter(log, service, hookConfig)

	utils.StartServer(config, "contigus", appName, router.Build(),
		utils.VersionInfo{
			Version:    version,
			Buildstamp: buildstamp,
			Githash:    githash,
		},
	)

}
