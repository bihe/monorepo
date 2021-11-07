package store_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/app/store"
)

const userName = "USER"

var sites = map[string][]store.UserSiteEntity{
	userName: {
		{
			Name:      "A",
			User:      userName,
			URL:       "http://urlA",
			PermList:  "a,b,c",
			CreatedAt: time.Now(),
		},
	},
}

func Test_New_Repo(t *testing.T) {
	repo := store.NewMock(sites)
	assert.NotNil(t, repo)
}

func Test_InUnitOfWork(t *testing.T) {
	repo := store.NewMock(sites)
	err := repo.InUnitOfWork(func(r store.Repository) error {
		if r == nil {
			return fmt.Errorf("no repo supplied")
		}
		return nil
	})
	assert.NoError(t, err)
}

func Test_Get_Sites_For_User(t *testing.T) {
	repo := store.NewMock(sites)
	// std.call
	s, err := repo.GetSitesForUser(userName)
	assert.NoError(t, err)
	assert.Len(t, s, 1)
	assert.Equal(t, "A", s[0].Name)

	// unknown username
	sites, err := repo.GetSitesForUser("userName")
	assert.NoError(t, err)
	assert.Empty(t, sites)
}

func Test_Store_Sites_For_User(t *testing.T) {
	repo := store.NewMock(sites)

	// std.call
	err := repo.StoreSiteForUser([]store.UserSiteEntity{
		{
			Name:      "B",
			User:      userName + "1",
			URL:       "http://urlB",
			PermList:  "b",
			CreatedAt: time.Now(),
		},
		{
			Name:      "B",
			User:      userName,
			URL:       "http://urlB",
			PermList:  "a,b",
			CreatedAt: time.Now(),
		},
	})
	assert.NoError(t, err)

	// get the number of users for the site from above
	users, err := repo.GetUsersForSite("B")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))
	assert.Contains(t, users, userName)
	assert.Contains(t, users, userName+"1")

	// get the number of users for the site from above
	users, err = repo.GetUsersForSite("C")
	assert.NoError(t, err)
	assert.Len(t, users, 0)
}

func Test_Store_Login(t *testing.T) {
	repo := store.NewMock(sites)

	err := repo.StoreLogin(store.LoginsEntity{
		User:      userName,
		Type:      store.DIRECT,
		CreatedAt: time.Now(),
	})
	assert.NoError(t, err)

	// second call - existing entry
	err = repo.StoreLogin(store.LoginsEntity{
		User:      userName,
		Type:      store.DIRECT,
		CreatedAt: time.Now(),
	})
	assert.NoError(t, err)

	// std.call
	count, err := repo.GetLoginsForUser(userName)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// user without logins
	count, err = repo.GetLoginsForUser(userName + "1")
	assert.Error(t, err)
}