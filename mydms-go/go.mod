module github.com/bihe/mydms

go 1.13

require (
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/aws/aws-sdk-go v1.29.24
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-chi/chi v4.1.0+incompatible
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/labstack/echo/v4 v4.1.15
	github.com/markusthoemmes/goautoneg v0.0.0-20190713162725-c6008fefa5b1
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1
	github.com/swaggo/echo-swagger v0.0.0-20191205130555-62f81ea88919
	github.com/swaggo/swag v1.6.5
	//golang.binggl.net/commons v1.0.14
	golang.binggl.net/commons v0.0.0
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	gopkg.in/ini.v1 v1.55.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace golang.binggl.net/commons => ../commons-go
