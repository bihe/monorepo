package appinfo

import (
	"net/http"

	"github.com/bihe/mydms/internal"
	"github.com/bihe/mydms/internal/security"
	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

// AppInfo provides information of the authenticated user and application meta-data
type AppInfo struct {
	UserInfo    UserInfo    `json:"userInfo"`
	VersionInfo VersionInfo `json:"versionInfo"`
}

// UserInfo provides information about authenticated user
type UserInfo struct {
	// DisplayName of authenticated user
	DisplayName string `json:"displayName"`
	// UserID of authenticated user
	UserID string `json:"userId"`
	// UserName of authenticated user
	UserName string `json:"userName"`
	// Email of authenticated user
	Email string `json:"email"`
	// Roles the authenticated user possesses
	Roles []string `json:"roles"`
}

// Claim defines a permission information for a given URL containing a specific role
type Claim struct {
	// Name of the application
	Name string `json:"name"`
	// URL of the application
	URL string `json:"url"`
	// Role as a form of permission
	Role string `json:"rol"`
}

// VersionInfo contains application meta-data
type VersionInfo struct {
	// Version of the application
	Version string `json:"version"`
	// BuildNumber defines the specific build
	BuildNumber string `json:"buildNumber"`
}

// Handler provides methos for applicatin metadata
type Handler struct {
	internal.VersionInfo
}

// GetAppInfo godoc
// @Summary provides information about the application
// @Description meta-data of the application including authenticated user and version
// @Tags appinfo
// @Produce  json
// @Success 200 {object} appinfo.AppInfo
// @Failure 401 {object} errors.ProblemDetail
// @Failure 403 {object} errors.ProblemDetail
// @Router /api/v1/appinfo [get]
func (h *Handler) GetAppInfo(c echo.Context) error {
	sc := c.(*security.ServerContext)
	id := sc.Identity
	log.Infof("Got user: %s", id.Username)

	a := AppInfo{
		UserInfo: UserInfo{
			DisplayName: id.DisplayName,
			UserID:      id.UserID,
			UserName:    id.Username,
			Email:       id.Email,
			Roles:       id.Roles,
		},
		VersionInfo: VersionInfo{
			Version:     h.Version,
			BuildNumber: h.Build,
		},
	}
	return c.JSON(http.StatusOK, a)
}
