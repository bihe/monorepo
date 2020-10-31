package appinfo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

func Test_GetAppInfo(t *testing.T) {
	svc := appinfo.NewService(logging.NewNop(), "1.0.0", "20200901-local")
	if svc == nil {
		t.Fatal("could not crate a new service instance")
	}

	ai, err := svc.GetAppInfo(&security.User{
		DisplayName:   "displayname",
		Authenticated: true,
		Email:         "email",
		Token:         "token",
		UserID:        "userid",
		Username:      "username",
		Roles:         []string{"a", "b"},
	})
	if err != nil {
		t.Errorf("could not get appinfo: %v", err)
	}

	assert.Equal(t, "displayname", ai.UserInfo.DisplayName)
	assert.Equal(t, "email", ai.UserInfo.Email)
	assert.Equal(t, "userid", ai.UserInfo.UserID)
	assert.Equal(t, "username", ai.UserInfo.UserName)
	assert.Equal(t, []string{"a", "b"}, ai.UserInfo.Roles)
	assert.Equal(t, "1.0.0", ai.VersionInfo.Version)
	assert.Equal(t, "20200901-local", ai.VersionInfo.BuildNumber)
}
