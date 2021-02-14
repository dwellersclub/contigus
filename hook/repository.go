package hook

import (
	"encoding/json"
	"time"

	"github.com/dwellersclub/contigus/models"
)

//Repository Hook configs repository
type Repository interface {
	GetHooks() []models.HookConfig
	addChangeListener(func([]models.HookConfig))
	Close() error
}

type basedRepo struct {
	listeners     []func([]models.HookConfig)
	latestVersion map[models.UUID]time.Time
	ticker        *time.Ticker
	cached        []models.HookConfig
}

func (br *basedRepo) decode(data []byte) (*models.HookConfig, error) {
	config := models.HookConfig{}
	err := json.Unmarshal(data, &config)
	return &config, err
}

func (br *basedRepo) addVersion(ID models.UUID, date time.Time) {
	br.latestVersion[ID] = date
}

func (br *basedRepo) addChangeListener(listener func([]models.HookConfig)) {
	br.listeners = append(br.listeners, listener)
}

func (br *basedRepo) IsNew(ID models.UUID) bool {
	_, ok := br.latestVersion[ID]
	return !ok
}

func (br *basedRepo) HasChanged(ID models.UUID, date time.Time) bool {
	prevDate, ok := br.latestVersion[ID]
	if ok {
		return !prevDate.Equal(date)
	}
	return false
}

func (br *basedRepo) startRefresh(seconds int64) <-chan time.Time {
	if br.ticker != nil {
		return br.ticker.C
	}
	br.ticker = time.NewTicker(time.Duration(seconds) * time.Second)
	return br.ticker.C
}

func (br *basedRepo) stopRefresh() {
	if br.ticker != nil {
		br.ticker.Stop()
		br.ticker = nil
	}
}

func (br *basedRepo) Close() error {
	if br.ticker != nil {
		br.stopRefresh()
	}
	return nil
}
