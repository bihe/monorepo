package store

import "fmt"

// MockRepo implements the repository but holds entries only in memory
// this is a defaulf implementation which can be used for testing
type MockRepo struct {
	sites  map[string][]UserSiteEntity
	logins map[string]int64
}

// compile guard for interface
var _ Repository = &MockRepo{}

func NewMock(sites map[string][]UserSiteEntity) Repository {
	return &MockRepo{
		sites:  sites,
		logins: make(map[string]int64),
	}
}

func (m *MockRepo) InUnitOfWork(fn func(repo Repository) error) error {
	return fn(m)
}

func (m *MockRepo) GetSitesForUser(user string) ([]UserSiteEntity, error) {
	if _, ok := m.sites[user]; !ok {
		return nil, fmt.Errorf("no sites for user available")
	}
	return m.sites[user], nil
}

func (m *MockRepo) GetUsersForSite(site string) ([]string, error) {
	var users []string
	var foundUsers map[string]bool = make(map[string]bool)
	for _, s := range m.sites {
		for _, u := range s {
			if u.Name == site {
				if _, ok := foundUsers[u.Name]; !ok {
					users = append(users, u.Name)
					foundUsers[u.Name] = true
				}
			}
		}
	}
	if len(users) == 0 {
		users = make([]string, 0)
	}
	return users, nil
}

func (m *MockRepo) GetLoginsForUser(user string) (int64, error) {
	if _, ok := m.logins[user]; !ok {
		return 0, fmt.Errorf("no login for user available")
	}
	return m.logins[user], nil
}

func (m *MockRepo) StoreLogin(login LoginsEntity) (err error) {
	var loginCount int64

	if _, ok := m.logins[login.User]; ok {
		loginCount = m.logins[login.User]
	}
	loginCount += 1
	m.logins[login.User] = loginCount
	return nil
}

func (m *MockRepo) StoreSiteForUser(sites []UserSiteEntity) (err error) {
	user := sites[0].User
	m.sites[user] = sites
	return nil
}
