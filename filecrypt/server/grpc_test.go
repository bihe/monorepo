package server

import (
	"testing"
)

func Test_ReadConfig(t *testing.T) {
	hostname, port, basePath, appConfig := readConfig()
	if hostname != "localhost" {
		t.Errorf("could not get hostname paramter")
	}
	if port != 3000 {
		t.Errorf("could not get port paramter")
	}
	if basePath != "./" {
		t.Errorf("could not get basepath paramter")
	}
	if appConfig.ServiceName != "testService" {
		t.Errorf("could not get serviceName")
	}
}
