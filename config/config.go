package config

import "leblanc.io/open-go-base/appconf"

type Config struct {
	Auth Auth `yaml:"auth"`

	Server Server `yaml:"server" env-prefix:"OGS_"`

	Database Database `yaml:"database" env-prefix:"OGS_"`

	Log appconf.Logging `yaml:"logging" env-prefix:"OGS_"`

	QRCode QRCode `yaml:"qrcode"`
}

// QRCode is the configuration of the QR codes of the short links.
type QRCode struct {
	// Logo is the path of a PNG or JPEG image drawn in the center of every QR
	// code. When empty, the QR codes are plain.
	Logo string `yaml:"logo" env:"OGS_QRCODE_LOGO" env-description:"Path of a PNG or JPEG image drawn in the center of the QR codes, plain QR codes if empty"`
}

// DefaultPasswordHash is the bcrypt hash of the default password
// "open-go-shorten".
const DefaultPasswordHash = "$2a$10$Dd172CerAZ/UvlEzt0mESORk6XmEMgQea1TRpKdyr7t6Xjayiyh/m"

type Auth struct {
	Username  string `yaml:"username" env:"OGS_USERNAME" env-default:"open-go-shorten" env-description:"Login of the administrator"`
	Password  string `yaml:"password" env:"OGS_PASSWORD" env-default:"$2a$10$Dd172CerAZ/UvlEzt0mESORk6XmEMgQea1TRpKdyr7t6Xjayiyh/m" env-description:"Bcrypt hash of the administrator password"`
	JwtSecret string `yaml:"jwt_secret" env:"OGS_JWT_SECRET" env-description:"Secret used to sign the authentication tokens, random if empty (sessions are then lost on restart)"`
}

// Server is the web server configuration: OGS_HOST, OGS_PORT, OGS_RATE_LIMIT,
// OGS_TRUSTED_PROXIES and OGS_PUBLIC_URL.
//
// It mirrors appconf.Web, whose default port (8080) can't be overridden: the
// default port of OpenGoShorten stays 9990.
type Server struct {
	Host           string   `yaml:"host"            env:"HOST"            env-default:"127.0.0.1" env-description:"Listen IP for web server"`
	Port           int      `yaml:"port"            env:"PORT"            env-default:"9990"      env-description:"Listen port"`
	RateLimit      int      `yaml:"rate_limit"      env:"RATE_LIMIT"      env-default:"100"       env-description:"Max requests per minute per client"`
	TrustedProxies []string `yaml:"trusted_proxies" env:"TRUSTED_PROXIES" env-separator:","       env-description:"Trusted proxy CIDRs"`

	// PublicURL is the base URL of the short links (e.g. https://s.example.com).
	// When empty, it is deduced from the incoming request.
	PublicURL string `yaml:"public_url" env:"PUBLIC_URL" env-description:"Base URL of the short links, deduced from the request if empty"`
}

// Database is the Redis configuration: OGS_REDIS_HOST, OGS_REDIS_PORT,
// OGS_REDIS_PASSWORD, OGS_REDIS_DB and OGS_DB_PREFIX.
type Database struct {
	appconf.Redis `yaml:",inline"`

	Prefix string `yaml:"prefix" env:"DB_PREFIX" env-default:"ogs-" env-description:"Prefix of the Redis keys"`
}
