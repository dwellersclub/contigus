package hook

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/dwellersclub/contigus/models"
	"github.com/sirupsen/logrus"
)

//NewFileBasedRepo Creates a new hook repo backed by a filesystem
func NewFileBasedRepo(path string, frequency int64) Repository {

	repo := fileBasedRepo{
		basedRepo: basedRepo{
			latestVersion: make(map[models.UUID]time.Time),
		},
		hooks: []models.HookConfig{},
		path:  path,
	}

	repo.addChangeListener(repo.updateHooks)

	refresher := repo.startRefresh(frequency)
	repo.refresh()

	go func() {
		for {
			select {
			case t := <-refresher:
				logrus.Infof("updating configuration: %s", t)
				repo.refresh()
			}
		}
	}()

	return &repo
}

type fileBasedRepo struct {
	basedRepo
	path  string
	hooks []models.HookConfig
}

func (fbr *fileBasedRepo) GetHooks() []models.HookConfig {
	return fbr.hooks
}

func (fbr *fileBasedRepo) updateHooks(items []models.HookConfig) {
	fbr.hooks = items
}

func (fbr *fileBasedRepo) _getHooks() ([]models.HookConfig, error) {
	var hooks []models.HookConfig
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if !info.IsDir() {

				data, readErr := ioutil.ReadFile(path)
				if readErr != nil {
					logrus.WithError(readErr).Errorf("can't read file config [%s]", path)
					return readErr
				}

				config, decodeErr := fbr.decode(data)
				if decodeErr != nil {
					logrus.WithError(decodeErr).Errorf("can't decode file config [%s]", path)
					return decodeErr
				}

				config.Date = info.ModTime()

				if valErr := config.Validate(); valErr != nil {
					logrus.WithError(valErr).Errorf("invalid file config [%s]", path)
					return valErr
				}

				hooks = append(hooks, *config)
			}
		} else {
			logrus.WithError(err).Errorf("can't read [%s]", path)
		}
		return nil
	}

	err := filepath.Walk(fbr.path, walkFn)
	if err != nil {
		logrus.WithError(err).Errorf("can't read config folder [%s]", fbr.path)
		return []models.HookConfig{}, err
	}

	return hooks, nil
}

func (fbr *fileBasedRepo) AddHook(models.HookConfig) error {
	return nil
}

func (fbr *fileBasedRepo) UpdateHook(models.HookConfig) error {
	return nil
}

func (fbr *fileBasedRepo) refresh() {
	hooks, err := fbr._getHooks()

	if err != nil {
		return
	}

	newHooks := []models.HookConfig{}

	foundIDs := make(map[models.UUID]bool)

	for _, hook := range hooks {
		foundIDs[hook.ID] = true
		if !fbr.IsNew(hook.ID) {
			if fbr.HasChanged(hook.ID, hook.Date) {
				newHooks = append(newHooks, hook)
			}
		} else {
			newHooks = append(newHooks, hook)
		}
	}

	for key := range fbr.latestVersion {
		ok := foundIDs[key]
		if !ok {
			newHooks = append(newHooks, models.HookConfig{
				ID:      key,
				Deleted: true,
			})
		}
	}

	for _, listener := range fbr.listeners {
		listener(newHooks)
	}

	for _, newHook := range newHooks {
		if !newHook.Deleted {
			fbr.latestVersion[newHook.ID] = newHook.Date
		} else {
			delete(fbr.latestVersion, newHook.ID)
		}
	}

}
