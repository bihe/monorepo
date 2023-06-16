package handler

// --------------------------------------------------------------------------
// Objects
// --------------------------------------------------------------------------

// Meta specifies application metadata
// swagger:model
type Meta struct {
	UserInfo UserInfo    `json:"userInfo"`
	Version  VersionInfo `json:"versionInfo"`
}

// UserInfo provides information about the currently logged-in user
// swagger:model
type UserInfo struct {
	DisplayName string   `json:"displayName"`
	UserID      string   `json:"userId"`
	UserName    string   `json:"userName"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
}

// VersionInfo is used to provide version and build
// swagger:model
type VersionInfo struct {
	Version string `json:"version"`
	Build   string `json:"buildNumber"`
}
