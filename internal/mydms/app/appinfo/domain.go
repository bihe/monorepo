package appinfo

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

// VersionInfo contains application meta-data
type VersionInfo struct {
	// Version of the application
	Version string `json:"version"`
	// BuildNumber defines the specific build
	BuildNumber string `json:"buildNumber"`
}
