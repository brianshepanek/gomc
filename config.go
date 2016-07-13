package gomc

type DatabaseConfig struct {
	UseDatabase string
	Username string
	Password string
	UseSSL bool
	Host string
	Port string
	Database string
	Table string
	Limit int
	Type string
}

type AppConfig struct {
	RequestValidateModel AppModel
	RequestValidateData interface{}
	LimitNonUser int
	LimitUser int
	RateLimitDataUseDatabaseConfig string
	Databases map[string]DatabaseConfig
}

var Config AppConfig