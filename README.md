# OpenGoShorten

<p align="center">
<img src="https://raw.githubusercontent.com/leblanc-simon/open-go-shorten/main/static/img/apple-touch-icon.png">
</p>

OpenGoShorten is a URL shortening service written in Go. It allows you to create short URLs for long ones, and it also provides a simple web interface to manage your shortened URLs.

It was created as a test of an AI's ability to generate a codebase and a README for the associated codebase. The majority of the code and this README were written by the AI (https://chat.mistral.ai/). Later evolutions (QR codes, security review, web interface redesign) were made with Claude (https://claude.ai/).

## Features

- URL shortening
- JWT authentication
- Redis database for storing URLs and visits
- Simple web interface for managing URLs
- Supports expiration dates for URLs
- Statistics for URLs (total visits, unique visitors)
- QR code of each short link (SVG and 512x512 PNG)
- Supports multiple platforms (Linux, MacOS, Windows)

## Installation

1. Clone the repository:

```
git clone https://github.com/leblanc-simon/open-go-shorten.git
```

Go 1.26.3 or later is required.

2. Install the dependencies:

```
go mod download
```

3. Build the application:

```
go build -o open-go-shorten main.go
```

4. Run the application:

```
./open-go-shorten
```

## Configuration

The application relies on [open-go-base](https://github.com/leblanc-simon/open-go-base) for its configuration,
logging and rate limiting. It can be configured using a YAML configuration file or environment variables.
The configuration file is optional: if it does not exist, the configuration is read from the environment
variables, and the default values are used for the missing ones.

The configuration file should be named `config.yaml` and should be located in the same directory as the application, or indicate config filename with `-c [/your/path/config.yaml]`.

`--help` lists every environment variable with its default value, `--version` prints the version.

The configuration file should have the following structure:

```yaml
auth:
  username: open-go-shorten
  password: $2a$10$Dd172CerAZ/UvlEzt0mESORk6XmEMgQea1TRpKdyr7t6Xjayiyh/m
  jwt_secret: change-me-with-a-long-random-string  # random if empty: sessions are lost on restart

database:
  host: 127.0.0.1
  port: 6379
  password:
  db: 0
  prefix: ogs-

server:
  host: 127.0.0.1
  port: 9990
  rate_limit: 100          # max requests per minute per client, 0 to disable
  trusted_proxies: []      # CIDR of your reverse proxies, e.g. ["127.0.0.1/32"]
  public_url:              # e.g. https://s.example.com, deduced from the request if empty

qrcode:
  logo:                    # PNG or JPEG image drawn in the center of the QR codes, plain QR codes if empty

logging:
  level: error             # debug|info|warn|error
  format: text             # text|json
  source: false
```

The environment variables that can be used to configure the application are:

- `OGS_USERNAME`, `OGS_PASSWORD`, `OGS_JWT_SECRET`
- `OGS_REDIS_HOST`, `OGS_REDIS_PORT`, `OGS_REDIS_PASSWORD`, `OGS_REDIS_DB`, `OGS_DB_PREFIX`
- `OGS_HOST`, `OGS_PORT`, `OGS_RATE_LIMIT`, `OGS_TRUSTED_PROXIES`, `OGS_PUBLIC_URL`
- `OGS_LOG_LEVEL`, `OGS_LOG_FORMAT`, `OGS_LOG_SOURCE`
- `OGS_QRCODE_LOGO`

When `jwt_secret` is empty, a random secret is generated at startup (the sessions are then lost on restart).
The default administrator password is `open-go-shorten`: change it.

The password is a bcrypt hash, which can be generated with `htpasswd -bnBC 10 "" your-password | tr -d ':\n'`.

### Reverse proxy

The IP of the visitors (used for the statistics and the rate limiting) is read from the `X-Forwarded-For` /
`X-Real-IP` headers **only** when the connection comes from a proxy listed in `trusted_proxies`. Behind a
reverse proxy, set `trusted_proxies` to its address, otherwise every visitor gets the IP of the proxy.

### Upgrading from 1.x

The configuration changed when moving to open-go-base:

| Before | Now |
|---|---|
| `OGS_SERVER_HOST`, `OGS_SERVER_PORT` | `OGS_HOST`, `OGS_PORT` |
| `OGS_DB_HOST`, `OGS_DB_PORT`, `OGS_DB_PASSWORD`, `OGS_DB_DATABASE` | `OGS_REDIS_HOST`, `OGS_REDIS_PORT`, `OGS_REDIS_PASSWORD`, `OGS_REDIS_DB` |
| `database.database` (YAML) | `database.db` |
| `server.log_level` (YAML) | `logging.level` |
| default `jwt_secret` | random secret (the previous public default is ignored) |

## Usage

### Web Interface

The web interface can be accessed by opening a web browser and navigating to `http://localhost:9990`. The default username and password are `open-go-shorten` and `open-go-shorten`, respectively.

### QR codes

Each short link has a QR code, available from the web interface (QR icon) or directly:

- `GET /qrcode/{shortURL}/svg`: scalable SVG image
- `GET /qrcode/{shortURL}/png`: 512x512 PNG image

Add `?download=1` to download the image as a file. The QR code encodes the full short link: its base URL is
deduced from the request, or forced with the `public_url` setting (useful behind a reverse proxy).

By default, the QR codes are plain. Set `qrcode.logo` (or `OGS_QRCODE_LOGO`) to the path of a PNG or JPEG image
to draw it in the center of every QR code. The QR codes then use the highest error correction level, so they
stay readable despite the logo. The image is loaded at startup: restart the server after changing it.


## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue. If you would like to contribute code, please fork the repository and submit a pull request.

## Author

* Simon Leblanc <contact@leblanc-simon.eu>
* Mistral AI (https://chat.mistral.ai/)
* Claude by Anthropic (https://claude.ai/)

## License

[WTFPL](http://www.wtfpl.net/)
