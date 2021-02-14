package utils

import (
	"database/sql"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func NewDB(dbURL string, maxIdle int, maxOpen int, ttl int, waitTime int64) *sql.DB {

	content := strings.Split(dbURL, "|")

	if len(content) < 2 {
		logrus.Errorf("Invalid dsn")
		return nil
	}

	var db *sql.DB

	Retry(logrus.New(), time.Duration(waitTime)*time.Second, time.Duration(waitTime*120)*time.Second, func() error {
		newDB, err := sql.Open(content[0], content[1])

		if err != nil {
			logrus.WithError(err).Error("can't connect to db with error")
			return err
		}

		db = newDB

		return nil
	})

	if db != nil {
		db.SetMaxIdleConns(maxIdle)
		db.SetMaxOpenConns(maxOpen)
		db.SetConnMaxLifetime(time.Minute * time.Duration(ttl))

		var now bool
		err := db.QueryRow("SELECT 1=1").Scan(&now)

		if err == nil && now {

			if waitTime > 10 {
				opts := prometheus.GaugeOpts{Name: "pool_connection_total", Help: "Number of db connection", Namespace: "marky", Subsystem: "db"}
				dbConnectionCounter := prometheus.NewGaugeFunc(opts, func() float64 {
					return float64(db.Stats().OpenConnections)
				})

				prometheus.MustRegister(dbConnectionCounter)
			}
			return db
		}
		logrus.WithError(err).Errorf("can't connect to db")
	}

	return nil
}

//GetDBUrl Replace place holder for a DB url connection
func GetDBUrl(url string, username string, password string, appName string) string {

	//check db if a driver is defined
	if !strings.Contains(url, "|") {
		logrus.Errorf("Url [%s] is not valid, should be [driver]|[connection string]", url)
		return ""
	}

	text := strings.Replace(url, "[username]", username, -1)
	text = strings.Replace(text, "[password]", password, -1)
	text = strings.Replace(text, "[appname]", appName, -1)

	return text
}
