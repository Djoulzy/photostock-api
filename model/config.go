package model

type Globals struct {
	LogLevel         int
	StartLogging     bool
	Mode             string
	DBType           string
	HTTP_addr        string
	HTTP_port        string
	AbsoluteBankPath string
	ImportDir        string
	ChunkDir         string
	DBDriver         string
	AssetsDir        string
}

type Postgres struct {
	DBHost   string
	DBPort   string
	DBUser   string
	DBPasswd string
	DBName   string
}

type SQLite struct {
	DBPath string
}

type ConfigData struct {
	Globals
	Postgres
	SQLite
}
