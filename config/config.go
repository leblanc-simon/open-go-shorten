package config

type Config struct {
	Auth struct {
		Username  string `yaml:"username" env:"OGS_USERNAME" env-default:"open-go-shorten"`
		Password  string `yaml:"password" env:"OGS_PASSWORD" env-default:"$2a$10$Dd172CerAZ/UvlEzt0mESORk6XmEMgQea1TRpKdyr7t6Xjayiyh/m"`
		JwtSecret string `yaml:"jwt_secret" env:"OGS_JWT_SECRET" env-default:"019263ef-1532-7333-98d6-3f08007f99f2"`
	} `yaml:"auth"`

	Database struct {
		Host     string `yaml:"host" env:"OGS_DB_HOST" env-default:"127.0.0.1"`
		Port     int    `yaml:"port" env:"OGS_DB_PORT" env-default:"6379"`
		Password string `yaml:"password" env:"OGS_DB_PASSWORD"`
		Database int    `yaml:"database" env:"OGS_DB_DATABASE" env-default:"0"`
		Prefix   string `yaml:"prefix" env:"OGS_DB_PREFIX" env-default:"ogs-"`
	} `yaml:"database"`

	Server struct {
		Host     string `yaml:"host" env:"OGS_SERVER_HOST" env-default:"127.0.0.1"`
		Port     int    `yaml:"port" env:"OGS_SERVER_PORT" env-default:"9990"`
		LogLevel string `yaml:"log_level" env:"OGS_LOG_LEVEL" env-default:"error"`
	} `yaml:"server"`
}
