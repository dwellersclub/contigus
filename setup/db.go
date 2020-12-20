package setup

var statements = []string{
	`CREATE TABLE %s_version ( version varchar(20) , created TIMESTAMP , updated TIMESTAMP )`,
	`CREATE TABLE %s_config ( base_url varchar(255) , created TIMESTAMP, updated TIMESTAMP)`,
}

//DBType database engine
type DBType string

//DBTypes database engines
type DBTypes struct {
	Mysql    DBType
	Postgres DBType
	SQLite   DBType
	MongoDB  DBType
}

//DBTypesEnum  database engines enum
var DBTypesEnum = DBTypes{
	Mysql:    DBType("mysql"),
	Postgres: DBType("pg"),
	SQLite:   DBType("sqlite"),
	MongoDB:  DBType("mongo"),
}

//DBConfig database config
type DBConfig struct {
	Type     DBType
	Username string
	Password string
	Schema   string
	Host     string
	Port     string
}

type dbInstallerTask struct{}
