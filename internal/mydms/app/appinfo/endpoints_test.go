package appinfo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/pkg/security"
)

type mockAppInfoService struct {
	fail bool
}

func (m mockAppInfoService) GetAppInfo(user *security.User) (ai appinfo.AppInfo, err error) {
	if m.fail {
		return appinfo.AppInfo{}, fmt.Errorf("error")
	}
	return appinfo.AppInfo{
		UserInfo: appinfo.UserInfo{
			DisplayName: user.DisplayName,
		},
		VersionInfo: appinfo.VersionInfo{
			Version: "1.0",
		},
	}, nil
}

func Test_GetAppInfo_Endpoint(t *testing.T) {
	e := appinfo.MakeGetAppInfoEndpoint(mockAppInfoService{})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}

	r, err := e(context.TODO(), appinfo.GetAppInfoRequest{
		User: &security.User{
			DisplayName: "DisplayName",
		},
	})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(appinfo.GetAppInfoResponse)
	if !ok {
		t.Error("response is not of type 'appinfo.GetAppInfoResponse'")
	}
	assert.Equal(t, "DisplayName", resp.AppInfo.UserInfo.DisplayName)
	assert.Equal(t, "1.0", resp.AppInfo.VersionInfo.Version)
}

func Test_GetAppInfo_Endpoint_Fail(t *testing.T) {
	e := appinfo.MakeGetAppInfoEndpoint(mockAppInfoService{fail: true})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}

	r, err := e(context.TODO(), appinfo.GetAppInfoRequest{
		User: &security.User{
			DisplayName: "DisplayName",
		},
	})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(appinfo.GetAppInfoResponse)
	if !ok {
		t.Error("response is not of type 'appinfo.GetAppInfoResponse'")
	}
	if resp.Err == nil {
		t.Error("error expected")
	}
	eString := resp.Failed().Error()
	assert.NotEmpty(t, eString)
}
