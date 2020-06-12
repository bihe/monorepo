package appinfo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/mydms"
	"golang.binggl.net/monorepo/mydms/security"
	sec "golang.binggl.net/monorepo/pkg/security"
)

func TestGetAppInfo(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	u := sec.User{
		Username:      "test",
		Roles:         []string{"roleA"},
		Email:         "a.b@c.de",
		UserID:        "1",
		DisplayName:   "User B",
		Authenticated: true,
	}
	sc := &security.ServerContext{Context: c, Identity: u}
	v := mydms.VersionInfo{
		Build:   "1",
		Version: "2",
	}
	h := Handler{VersionInfo: v, Log: log.New().WithField("mode", "test")}

	// Assertions
	if assert.NoError(t, h.GetAppInfo(sc)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var ai AppInfo
		err := json.Unmarshal(rec.Body.Bytes(), &ai)
		if err != nil {
			t.Errorf("could not get valid json: %v", err)
		}
		assert.Equal(t, v.Build, ai.VersionInfo.BuildNumber)
		assert.Equal(t, v.Version, ai.VersionInfo.Version)

		assert.Equal(t, u.Username, ai.UserInfo.UserName)
		assert.Equal(t, u.UserID, ai.UserInfo.UserID)
		assert.Equal(t, u.Email, ai.UserInfo.Email)
		assert.Equal(t, u.DisplayName, ai.UserInfo.DisplayName)
		assert.Equal(t, u.Roles, ai.UserInfo.Roles)
	}
}
