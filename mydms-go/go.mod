module github.com/bihe/mydms

go 1.13

require (
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/aws/aws-sdk-go v1.29.24
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/labstack/echo/v4 v4.1.15
	github.com/markusthoemmes/goautoneg v0.0.0-20190713162725-c6008fefa5b1
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/swaggo/echo-swagger v0.0.0-20191205130555-62f81ea88919
	github.com/swaggo/swag v1.6.5
	//golang.binggl.net/commons v1.0.14
	golang.binggl.net/commons v0.0.0
)

replace (
	golang.binggl.net/commons => ../commons-go
)
