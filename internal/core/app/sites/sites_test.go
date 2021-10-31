package sites_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/pkg/security"
)

const adminRole = "ADMIN"
const userName = "USER"

var testSites []store.UserSiteEntity = []store.UserSiteEntity{
	{
		Name:      "A",
		User:      userName,
		URL:       "http://urlA",
		PermList:  adminRole + ",a,b,c",
		CreatedAt: time.Now(),
	},
}

func newMockRepo() store.Repository {
	svc := make(map[string][]store.UserSiteEntity)
	svc[userName] = testSites
	return store.NewMock(svc)
}

func Test_New_Sites(t *testing.T) {
	svc := sites.New(adminRole, newMockRepo())
	if svc == nil {
		t.Errorf("could not initialize sites.Service")
	}
}

func Test_Get_Sites_For_User(t *testing.T) {
	svc := sites.New(adminRole, newMockRepo())

	// std.call
	s, err := svc.GetSitesForUser(security.User{
		Username: userName,
		Email:    userName,
		Roles:    []string{adminRole, "a"},
	})
	if err != nil {
		t.Errorf("could not get sites: %v", err)
	}
	if len(s.Sites) != 1 {
		t.Errorf("could not get sites for user")
	}
	assert.True(t, s.Editable)

	// non-edit
	s, err = svc.GetSitesForUser(security.User{
		Username: userName,
		Email:    userName,
		Roles:    []string{"a"},
	})
	if err != nil {
		t.Errorf("could not get sites: %v", err)
	}
	if len(s.Sites) != 1 {
		t.Errorf("could not get sites for user")
	}
	assert.False(t, s.Editable)

	// sites not found
	s, err = svc.GetSitesForUser(security.User{
		Username: "userName",
		Email:    "userName",
	})
	assert.NoError(t, err)
	assert.Empty(t, s.Sites)
}

func Test_Get_Users_For_Site(t *testing.T) {
	svc := sites.New(adminRole, newMockRepo())

	// std.call
	users, err := svc.GetUsersForSite("A", security.User{
		Username: userName,
		Email:    userName,
		Roles:    []string{adminRole, "a"},
	})
	if err != nil {
		t.Errorf("could not get users for site; %v", err)
	}
	if len(users) != 1 {
		t.Errorf("returned users do not match")
	}

	// missing perm
	_, err = svc.GetUsersForSite("A", security.User{
		Username: userName,
		Email:    userName,
		Roles:    []string{"a"},
	})
	if err == nil {
		t.Errorf("error expected")
	}

	// wrong site
	_, err = svc.GetUsersForSite("B", security.User{
		Username: userName,
		Email:    userName,
		Roles:    []string{adminRole, "a"},
	})
	assert.NoError(t, err)
}

func Test_Save_Sites_For_User(t *testing.T) {
	svc := sites.New(adminRole, newMockRepo())

	// std.call
	var s sites.UserSites
	s.Sites = []sites.SiteInfo{
		{
			Name: "B",
			URL:  "http://urlB",
			Perm: []string{"b"},
		},
	}
	s.User = userName
	err := svc.SaveSitesForUser(s, security.User{
		Username: userName,
		Email:    userName,
		Roles:    []string{adminRole, "a"},
	})
	if err != nil {
		t.Errorf("error saving sites")
	}

	// missing sites
	err = svc.SaveSitesForUser(sites.UserSites{}, security.User{
		Username: userName,
		Email:    userName,
		Roles:    []string{adminRole, "a"},
	})
	if err == nil {
		t.Errorf("error expected")
	}

	// missing permission
	err = svc.SaveSitesForUser(sites.UserSites{}, security.User{
		Username: userName,
		Email:    userName,
		Roles:    []string{"a"},
	})
	if err == nil {
		t.Errorf("error expected")
	}
}
