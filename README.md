# OpenGoShorten

<p align="center">
<img src="https://raw.githubusercontent.com/leblanc-simon/open-go-shorten/main/static/img/apple-touch-icon.png">
</p>

OpenGoShorten is a URL shortening service written in Go. It allows you to create short URLs for long ones, and it also provides a simple web interface to manage your shortened URLs.

It was created as a test of an AI's ability to generate a codebase and a README for the associated codebase. The majority of the code and this README were written by the AI (https://chat.mistral.ai/).

## Features

- URL shortening
- JWT authentication
- Redis database for storing URLs and visits
- Simple web interface for managing URLs
- Supports expiration dates for URLs
- Statistics for URLs (total visits, unique visitors)
- Supports multiple platforms (Linux, MacOS, Windows)

## Installation

1. Clone the repository:

```
git clone https://github.com/leblanc-simon/open-go-shorten.git
```

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

The application can be configured using a YAML configuration file or environment variables. The configuration file is optional, and if it is not provided, the application will use the default values.

The configuration file should be named `config.yaml` and should be located in the same directory as the application, or indicate config filename with `-c [/your/path/config.yaml]`.

The configuration file should have the following structure:

```yaml
auth:
  username: open-go-shorten
  password: $2a$10$Dd172CerAZ/UvlEzt0mESORk6XmEMgQea1TRpKdyr7t6Xjayiyh/m
  jwt_secret: 019263ef-1532-7333-98d6-3f08007f99f2

database:
  host: 127.0.0.1
  port: 6379
  password:
  database: 0
  prefix: ogs-

server:
  host: 127.0.0.1
  port: 9990
  log_level: error
```

The environment variables that can be used to configure the application are:

- `OGS_USERNAME`
- `OGS_PASSWORD`
- `OGS_JWT_SECRET`
- `OGS_DB_HOST`
- `OGS_DB_PORT`
- `OGS_DB_PASSWORD`
- `OGS_DB_DATABASE`
- `OGS_DB_PREFIX`
- `OGS_SERVER_HOST`
- `OGS_SERVER_PORT`
- `OGS_LOG_LEVEL`

## Usage

### Web Interface

The web interface can be accessed by opening a web browser and navigating to `http://localhost:9990`. The default username and password are `open-go-shorten` and `open-go-shorten`, respectively.


## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue. If you would like to contribute code, please fork the repository and submit a pull request.

## Author

* Simon Leblanc <contact@leblanc-simon.eu>
* Mistral AI (https://chat.mistral.ai/)

## License

[WTFPL](http://www.wtfpl.net/)
