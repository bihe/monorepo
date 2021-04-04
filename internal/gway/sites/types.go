package sites

// UserSites holds information about the current user and sites
type UserSites struct {
	User     string     `json:"user"`
	Editable bool       `json:"editable"`
	Sites    []SiteInfo `json:"userSites"`
}

// SiteInfo holds data of a site
type SiteInfo struct {
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Perm []string `json:"permissions"`
}

// UserList holds the usernames for a given site
type UserList struct {
	Count int      `json:"count"`
	Users []string `json:"users"`
}
