module github.com/bihe/bookmarks

go 1.13

require (
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-chi/render v1.0.1
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/jinzhu/gorm v1.9.12
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/wangii/emoji v0.0.0-20150519084846-d15b69a4831e
	golang.binggl.net/commons v0.0.0
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace golang.binggl.net/commons => ../commons-go
