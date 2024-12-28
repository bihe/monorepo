package sites

import (
	"fmt"
	"strings"

	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/pkg/security"
)

// --------------------------------------------------------------------------
// Interface definition
// --------------------------------------------------------------------------

// Service defines the logic to access information about sites/permissions of
// a specific user
type Service interface {
	// GetSitesForUser returns the associated sites and permission for the given user
	GetSitesForUser(user security.User) (UserSites, error)
	// GetUsersForSite takes the name of a site and returns the users who have access to the given site
	// the provided user is used to check if the user has access to retrieve this information
	GetUsersForSite(site string, user security.User) ([]string, error)
	// SaveSitesForUsers takes the provided sites and saves the object. Either for the same user
	// or for other users as well, if the supplied user has the necessary permissions to perform this action
	SaveSitesForUser(sites UserSites, user security.User) error
}

// New creates a Service instance. The admin-role to check for privileged/admin activities is specified
func New(adminRole string, repository store.Repository) Service {
	return &siteService{
		adminRole: adminRole,
		repo:      repository,
	}
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

var _ Service = &siteService{}

type siteService struct {
	adminRole string
	repo      store.Repository
}

func (s *siteService) GetSitesForUser(user security.User) (UserSites, error) {
	sites, err := s.repo.GetSitesForUser(user.Email)
	if err != nil {
		return UserSites{}, fmt.Errorf("could not get sites for user: %s; %v", user.Email, err)
	}

	u := UserSites{
		User:     user.Username,
		Sites:    make([]SiteInfo, 0),
		Editable: hasRole(user, s.adminRole),
	}
	for _, s := range sites {
		parts := strings.Split(s.PermList, security.RoleDelimiter)
		u.Sites = append(u.Sites, SiteInfo{
			Name: s.Name,
			URL:  s.URL,
			Perm: parts,
		})
	}
	return u, nil
}

func (s *siteService) GetUsersForSite(site string, user security.User) ([]string, error) {
	if !hasRole(user, s.adminRole) {
		return nil, fmt.Errorf("not allowed to perform this action, missing admin-role for user '%s'", user.Email)
	}
	sites, err := s.repo.GetUsersForSite(site)
	if err != nil {
		return nil, fmt.Errorf("could not get users for site %s; %v", site, err)
	}
	return sites, nil
}

func (s *siteService) SaveSitesForUser(sites UserSites, user security.User) error {
	var storeEntities []store.UserSiteEntity = make([]store.UserSiteEntity, 0)
	if !hasRole(user, s.adminRole) {
		return fmt.Errorf("not allowed to perform this action, missing admin-role for user '%s'", user.Email)
	}
	if len(sites.Sites) == 0 {
		return fmt.Errorf("no sites supplied to store")
	}
	for _, s := range sites.Sites {
		storeEntities = append(storeEntities, store.UserSiteEntity{
			Name:     s.Name,
			URL:      s.URL,
			User:     user.Email,
			PermList: strings.Join(s.Perm, security.RoleDelimiter),
		})
	}
	err := s.repo.InUnitOfWork(func(repo store.Repository) error {
		return repo.StoreSiteForUser(storeEntities)
	})
	if err != nil {
		return fmt.Errorf("could not store the supplied sites; %v", err)
	}
	return nil
}

// hasRole checks if the given user has the given role
func hasRole(user security.User, role string) bool {
	for _, p := range user.Roles {
		if p == role {
			return true
		}
	}
	return false
}
