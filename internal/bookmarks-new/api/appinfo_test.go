package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/api"
	"golang.binggl.net/monorepo/pkg/handler"
)

func appinfoHandler() http.Handler {
	return handlerWith(&handlerOps{
		version: "1.0",
		build:   "today",
	})
}

func Test_GetAppInfo(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/bookmarks/appinfo", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	appinfoHandler().ServeHTTP(rec, req)

	// assert
	var appInfo handler.Meta
	assert.Equal(t, 200, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &appInfo); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, "1.0", appInfo.Version.Version)
	assert.Equal(t, "today", appInfo.Version.Build)
}

func Test_GetWhoAmI(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/whoami", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	appinfoHandler().ServeHTTP(rec, req)

	// assert
	var whoAmI api.WhoAmIResponse
	assert.Equal(t, 200, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &whoAmI); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, "User A", whoAmI.DisplayName)
	assert.Equal(t, "user@a.com", whoAmI.Email)
}
