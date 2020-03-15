module github.com/bihe/onefrontend

go 1.13

require (
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/kr/pretty v0.2.0 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/wangii/emoji v0.0.0-20150519084846-d15b69a4831e
	golang.binggl.net/commons v0.0.0
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace golang.binggl.net/commons => ../commons-go
