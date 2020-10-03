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
