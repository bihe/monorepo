package store

// --------------------------------------------------------------------------
// Repository interface
// --------------------------------------------------------------------------

// Repository defines the methods interacting with the store
type Repository interface {
	GetSitesForUser(user string) ([]UserSiteEntity, error)
	GetUsersForSite(site string) ([]string, error)
	StoreSiteForUser(sites []UserSiteEntity) (err error)
}
