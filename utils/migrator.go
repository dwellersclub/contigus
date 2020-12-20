package utils

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	logging "github.com/dwellersclub/contigus/log"
)

//NewMigrator create a new instance of a migrator
func NewMigrator(paths []string, dbURL string) (*Migrator, error) {
	logger := logging.GetLogger().WithField("module", "migrator")

	sqls, err := loadMigrations(logger, paths)
	if err != nil {
		return nil, err
	}

	db := NewDB(dbURL, 1, 1, 1, 10)
	if db == nil {
		return nil, fmt.Errorf("can't create db")
	}

	return &Migrator{sqls: sqls, sqlEngine: sqlx.NewDb(db, "postgres"), Logger: logger}, nil
}

//Migrator Sql migrator
type Migrator struct {
	sqls      []sqlContent
	sqlEngine *sqlx.DB
	Logger    *logrus.Entry
}

//Migrate Apply applies migration scripts
func (mg *Migrator) Migrate() error {

	migrations, err := mg.getMigrations() // To apply schema if missing

	if err != nil {
		return err
	}

	if len(migrations) > 0 {
		lastMigration := migrations[len(migrations)-1]
		mg.Logger.Infof("Last migration ran | %d - %s / %s [%s]", lastMigration.Version, lastMigration.Package, lastMigration.Name, lastMigration.CreatedDate)
	}

	for _, sqlContent := range mg.sqls {

		_, err := mg.getMigration(sqlContent.Version, sqlContent.Name, sqlContent.Package)

		if err != nil {

			if err != sql.ErrNoRows {
				return err
			}

			err = mg.applyMigration(sqlContent)
			mg.Logger.Infof("Apply migration version [%d] - [%s] / [%s]", sqlContent.Version, sqlContent.Package, sqlContent.Name)
			if err != nil {
				mg.Logger.Infof("Rollback migration version [%d] - [%s] / [%s]", sqlContent.Version, sqlContent.Package, sqlContent.Name)
				mg.applyMigrationRollback(sqlContent)
				return err
			}
		}

	}
	return nil
}

func (mg *Migrator) applyMigration(sqlMigration sqlContent) error {
	tx, err := mg.sqlEngine.Begin()

	if err != nil {
		return err
	}

	statements := strings.Split(sqlMigration.UpContent, ";")

	for _, statement := range statements {
		_, err := tx.Exec(statement)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Can't apply migration %s /n/n %s", err.Error(), statement)
		}
	}

	_, err = tx.Exec("INSERT INTO migrator (name, version, package) VALUES ($1, $2, $3) ", sqlMigration.Name, sqlMigration.Version, sqlMigration.Package)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s", err.Error())
	}

	tx.Commit()

	return nil
}

func (mg *Migrator) applyMigrationRollback(sqlMigration sqlContent) error {
	tx, err := mg.sqlEngine.Begin()

	if err != nil {
		return err
	}

	statements := strings.Split(sqlMigration.DownContent, ";")

	for _, statement := range statements {
		_, err := tx.Exec(statement)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Can't apply rollback %s /n/n %s", err.Error(), statement)
		}
	}

	tx.Commit()

	return nil
}

func (mg *Migrator) getMigrations() ([]dbMigration, error) {
	migrations := []dbMigration{}
	err := mg.sqlEngine.Select(&migrations, "SELECT * FROM migrator order by created_date asc, version asc")
	if err != nil {

		schema := `CREATE TABLE migrator (
			checksum varchar(255),
			version integer ,
			package varchar(255) ,
			name varchar(255),
			created_date timestamp DEFAULT CURRENT_TIMESTAMP,
    	CONSTRAINT migrator_pkey PRIMARY KEY (version, package, name)
		);`

		//create table
		_, err := mg.sqlEngine.Exec(schema)

		return []dbMigration{}, err
	}
	return migrations, nil
}

func (mg *Migrator) getMigration(version int, name string, pkg string) (*dbMigration, error) {
	migration := dbMigration{}
	err := mg.sqlEngine.Get(&migration, "SELECT * FROM migrator WHERE name=$1 and version=$2 and package=$3", name, version, pkg)
	if err != nil {
		return nil, err
	}
	return &migration, nil
}

func loadMigrations(logger *logrus.Entry, paths []string) ([]sqlContent, error) {

	sqlContentMapped := make(map[string]*sqlContent)

	var err error

	for _, path := range paths {
		var files []os.FileInfo
		files, err = ioutil.ReadDir(path)

		if err != nil {
			logger.WithError(err).Warnf("can't find folder %s", path)
			continue
		}

		var abs string
		abs, err = filepath.Abs(path)
		if err != nil {
			logger.WithError(err).Errorf("can't read files from folder %s", path)
			continue
		}

		if err != nil {
			return nil, err
		}

		pckgName := filepath.Base(filepath.Dir(abs))

		for _, file := range files {
			if !file.IsDir() {
				tokens := strings.Split(file.Name(), "_")
				if len(tokens) == 3 {
					rawVersion := tokens[0]

					i, err := strconv.Atoi(rawVersion)
					if err != nil {
						return nil, err
					}

					name := tokens[1]
					runType := tokens[2]

					key := fmt.Sprintf("%s_%s_%s", rawVersion, path, name)

					mig, ok := sqlContentMapped[key]

					if !ok {
						mig = &sqlContent{
							Version: i,
							Name:    name,
							Package: pckgName,
						}
						sqlContentMapped[key] = mig
					}

					bytes, err := ioutil.ReadFile(path + "/" + file.Name())
					if err != nil {
						return nil, err
					}

					if strings.Compare(runType, "down") == 0 {
						mig.DownContent = string(bytes)
					} else {
						mig.UpContent = string(bytes)
					}
				}
			}
		}
	}

	var sqls []sqlContent
	for _, sqlContent := range sqlContentMapped {
		sqls = append(sqls, *sqlContent)
	}

	sort.SliceStable(sqls, func(i, j int) bool {
		return sqls[i].Version < sqls[j].Version
	})

	logger.Infof("Found [%d] migration script(s)", len(sqls))

	return sqls, nil
}

type sqlContent struct {
	Version     int
	Name        string
	UpContent   string
	DownContent string
	Package     string
}

func (sqlc *sqlContent) Checksum() string {
	return ""
}

type dbMigration struct {
	Version     int       `db:"version"`
	Name        string    `db:"name"`
	Checksum    *string   `db:"checksum"`
	Package     string    `db:"package"`
	CreatedDate time.Time `db:"created_date"`
}
