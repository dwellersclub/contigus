package main

import (
	"os"

	"github.com/caarlos0/env"

	"github.com/dwellersclub/contigus/hook"
	logging "github.com/dwellersclub/contigus/log"
	"github.com/dwellersclub/contigus/utils"
)

var log = logging.GetLogger()
var (
	version    string
	buildstamp string
	githash    string
)

func main() {
	var appName = "hook"

	config := utils.ServerConfig{}
	err := env.Parse(&config)

	if err != nil {
		log.WithError(err).Errorf("Can't start %s", appName)
		os.Exit(1)
		return
	}

	router := hook.NewRouter(log)

	utils.StartServer(config, "contigus", appName, router.Build(),
		utils.VersionInfo{Version: version, Buildstamp: buildstamp, Githash: githash})

}
