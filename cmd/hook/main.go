package main

import (
	"os"

	"github.com/caarlos0/env"
	_ "github.com/lib/pq"

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

	config := utils.HookServerConfig{
		ServerConfig: utils.ServerConfig{
			Port: 8081,
		},
	}

	err := env.Parse(&config)

	if err != nil {
		log.WithError(err).Errorf("Can't start %s", appName)
		os.Exit(1)
		return
	}

	repo := hook.NewFileBasedRepo(config.ConfigPath, config.RefreshFreq)

	encryptor := hook.NewEncryptor()
	metrics := hook.NewHookMetrics()

	service := hook.NewService(metrics, encryptor, "", nil, repo)

	router := hook.NewRouter(log, service, config, nil)

	utils.StartServer(config.ServerConfig, "contigus", appName, router.Build(),
		utils.VersionInfo{
			Version:    version,
			Buildstamp: buildstamp,
			Githash:    githash,
		},
	)

}
